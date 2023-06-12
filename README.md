# Cloud Platform Reports

[![Releases](https://img.shields.io/github/release/ministryofjustice/cloud-platform-how-out-of-date-are-we/all.svg?style=flat-square)](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/releases)

Various reports about the Cloud Platform, displayed in a web application: https://reports.cloud-platform.service.justice.gov.uk

See the [about page](views/about.erb) for more details.

## Tech overview

This repo contains 3 things:

* [Front-end web application "Cloud Platform Reports"](#front-end-web-application)
* [Scheduled jobs "Cronjobs"](#scheduled-jobs) that collect data
* ["Dashboard-reporter"](#dashboard-reporter) that sends reports to teams' Slack channels

## Front-end web application

"Cloud Platform Reports" is a simple Sinatra web application with database and API.

What it does:

* receives data from the scheduled jobs via its API
* stores the data in DynamoDB
* serves report web pages at https://reports.cloud-platform.service.justice.gov.uk
  * on each request, data is read from DynamoDB, processed, then fed into an HTML template

Deployment:

* Helm Chart is defined: [/cloud-platform-reports](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/tree/main/cloud-platform-reports)
* CI/CD:
  * CI - Docker images are built and put on DockerHub by a [GitHub Action workflow](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/blob/main/.github/workflows/docker-hub.yml)
  * CD - Chart is deployed manually into 'live' cluster in namespace such as `cloud-platform-reports-dev`

### Developing

If you have a working ruby 2.7 environment, you can run the web application locally as follows:

```sh
bundle install
./app.rb
```

### Data Storage

The web application currently has two options for backend data storage:

* Filestore: POSTed JSON data is stored/retrieved as files in the local filesystem, below the `data` directory.
* AWS DynamoDB: POSTed JSON data is stored/retrieved as documents in a DynamoDB table, where the key is the same filename that would be used if `Filestore` were the storage backend.

The application will use `Filestore` unless a `DYNAMODB_TABLE_NAME` environment variable is configured.

#### Using DyanamoDB storage

To use DynamoDB as the storage backend, the following environment variables must be set:

* `DYNAMODB_REGION`: e.g. "eu-west-2"
* `DYNAMODB_ACCESS_KEY_ID`: An AWS access key with permission to access the DynamoDB table
* `DYNAMODB_SECRET_ACCESS_KEY`: An AWS secret key corresponding to the access key
* `DYNAMODB_TABLE_NAME`: The name of the DynamoDB table - this should have a `filename` key field

### Updating the JSON data

In all cases, POSTing JSON data to `/endpoint` will result in the post body being stored as `data/endpoint.json`, provided the correct API key is provided in the `X-API-KEY` header.

JSON data should consist of a hash with at least two key/value pairs:

* `updated_at` containing a time value in a human-readable string format
* A named data structure (the name can be any string value), containing the bulk of the data comprising the report.

e.g. The report on MoJ Github repositories might consist of:

```json
{
    "updated_at": "2020-09-16 15:23:42 UTC",
    "repositories": [ ...list of data hashes, one for each repo...]
}
```

The app. will only accept posted JSON data when the HTTP POST supplies the correct API key.

'correct' means the value of the 'X-API-KEY' header in the HTTP POST must match the value of the 'API_KEY' environment variable that was in scope when the app. was started.

If the supplied API key matches the expected value, the locally stored JSON data file will be overwritten with the request body supplied in the POST.

If the API key doesn't match, the app. will return a 403 error.

## Scheduled jobs

What these do:

* Data for each report is collected by Docker images defined in [/reports](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/tree/main/reports)
* These get run on a schedule, using a k8s [CronJob](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/tree/main/cloud-platform-reports-cronjobs/templates)
* They POST the data as JSON to the front-end web application

Deployment:

* Helm Chart defined in [/cloud-platform-reports-cronjobs](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/tree/main/cloud-platform-reports-cronjobs)
* CI/CD:
  * CI - Docker images are built and put on DockerHub by a [GitHub Action workflow](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/blob/main/.github/workflows/docker-hub.yml)
  * CD - Chart is deployed manually into the Manager cluster in namespace `concourse-main`

## Install and update

This section describes how to install and update the front-end app and scheduled jobs on the server, using the two Helm Charts. (We haven't got round to creating an automated pipeline yet.)

### Pre-requisites

* access to perform `awscli` commands in the account `cloud-platform-aws`
* `live` kube context
* `manager` kube context
* `cloud-platform-reports-cronjobs/secrets.yaml` file containing Docker Hub credentials and API key in double encoded format
* `cloud-platform-reports/secrets.yaml` file defining the web application API key in encoded format

The web application API key is required by both the web application and the
cronjobs which post the report data. 

## API Key Secret

The web application deployment is now configured to pull the API key from AWS Secrets Manager; deployment/upgrade of the Helm chart in production namespace will fetch this value directly
without additional manual steps. The kubernetes secret for this value is managed via the
[Secrets Manager module](https://github.com/ministryofjustice/cloud-platform-environments/blob/main/namespaces/live.cloud-platform.service.justice.gov.uk/cloud-platform-reports-prod/resources/secret.tf), with the secret name [`hoodaw-api-key`](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/blob/main/cloud-platform-reports/values.yaml#L26).


When deploying the cronjob from `cloud-platform-reports-cronjobs`, the API key defined in the
`cloud-platform-reports-cronjobs/secrets.yaml` should be encoded again(double encoded from real API key). That way when the kubernetes secret is decoded from the cronjob and the API key still remains in encoded format when sending the data via the POST which then matches the value of the secret in `cloud-platform-reports/secrets.yaml`.

 Equivalent secrets are created in both the `live/<web app>` and `manager/concourse-main` namespaces.

### Install

```sh
make deploy
```

**NOTE** The cronjobs for both dev and prod are deployed into the same namespace (`manager/concourse-main`). So you need to change the chart name if you want to deploy development cronjobs alongside the existing production cronjobs.

### Updating

To update this application:

1. Make your code changes, PR them and merge the PR once approved
1. Create a new release [via the github UI](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/releases).
   This will trigger a GitHub action to build all the docker images (web application and scheduled jobs), and push them to docker hub tagged with the release. The `appVersion` is used in the template files so that the same version of the relevant docker image is used for each sub-component of the application.
1. Edit `cloud-platform-reports/Chart.yaml` and `cloud-platform-reports-cronjobs/Chart.yaml` - change `appVersion` to the new release number
1. Update the application by running:

```sh
make upgrade
```

## Dashboard Reporter

This script sends reports to teams' Slack channels.

* Code: [/dashboard-reporter](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/tree/main/dashboard-reporter)
* Runs every 24h by Concourse: [Concourse hoodaw-dashboard-reporter job](https://concourse.cloud-platform.service.justice.gov.uk/teams/main/pipelines/hoodaw/jobs/hoodaw-dashboard-reporter)

The `dashboard-reporter` directory maintains a script which will
generate a report, formatted for use as a slack message,
containing the information on the dashboard page of the web
application.

The code in the reporter script is built from classes defined in the main
project, purely so that we can keep the Dockerfile simple and just add a single
ruby script to the default ruby alpine image without having to install gems
etc.

---
last_reviewed_on: 2021-06-30
review_in: 3 months
---
