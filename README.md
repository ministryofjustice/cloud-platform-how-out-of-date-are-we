# Cloud Platform Reports

[![Releases](https://img.shields.io/github/release/ministryofjustice/cloud-platform-how-out-of-date-are-we/all.svg?style=flat-square)](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/releases)

Various reports about the Cloud Platform, [displayed via a web application.](https://reports.cloud-platform.service.justice.gov.uk)

See the [about page](views/about.erb) for more details.

## Installation

The application consists of 2 Helm charts - one for the [web application](cloud-platform-reports), and one for the [cronjobs](cloud-platform-reports-cronjobs) which provide the data.

The web application is deployed to the `live-1` cluster. The cronjobs are deployed to the `concourse-main` namespace of the `manager` cluster.

## Pre-requisites

* `live-1` kube context
* `manager` kube context
* `cloud-platform-reports-cronjobs/secrets.yaml` file containing Docker Hub credentials
* `cloud-platform-reports/secrets.yaml` file defining the web application API key

The web application API key is required by both the web application and the
cronjobs which post the report data. So the
`cloud-platform-reports/secrets.yaml` file is also used when deploying the
`cloud-platform-reports-cronjobs` helm chart. Equivalent secrets are created in
both the `live-1/<web app>` and `manager/concourse-main` namespaces.

## Deploying

```
make deploy
```

## Updating

To update this application:

1. Make your code changes, PR them and merge the PR once approved
1. Create a new release [via the github UI](https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/releases)
1. Edit `cloud-platform-reports/Chart.yaml` and `cloud-platform-reports-cronjobs/Chart.yaml` - change `appVersion` to the new release number
1. Update the application by running:

```
make upgrade
```

> When a new release is created, a github action rebuilds all the project's
> docker images, and tags them with the release number. The `appVersion` is
> used in the template files so that the same version of the relevant docker
> image is used for each sub-component of the application.

## Data Storage

The web application currently has two options for backend data storage:

* Filestore: POSTed JSON data is stored/retrieved as files in the local filesystem, below the `data` directory.
* AWS DynamoDB: POSTed JSON data is stored/retrieved as documents in a DynamoDB table, where the key is the same filename that would be used if `Filestore` were the storage backend.

The application will use `Filestore` unless a `DYNAMODB_TABLE_NAME` environment variable is configured.

### Using DyanamoDB storage

To use DynamoDB as the storage backend, the following environment variables must be set:

* `DYNAMODB_REGION`: e.g. "eu-west-2"
* `DYNAMODB_ACCESS_KEY_ID`: An AWS access key with permission to access the DynamoDB table
* `DYNAMODB_SECRET_ACCESS_KEY`: An AWS secret key corresponding to the access key
* `DYNAMODB_TABLE_NAME`: The name of the DynamoDB table - this should have a `filename` key field

## Dashboard Reporter

The `dashboard-reporter` directory maintains a script which will
generate a report, formatted for use as a slack message,
containing the information on the dashboard page of the web
application.

The code in the reporter script is built from classes defined in the main
project, purely so that we can keep the Dockerfile simple and just add a single
ruby script to the default ruby alpine image without having to install gems
etc.

## Updating the JSON data

In all cases, POSTing JSON data to `/endpoint` will result in the post body being stored as `data/endpoint.json`, provided the correct API key is provided in the `X-API-KEY` header.

JSON data should consist of a hash with at least two key/value pairs:
* `updated_at` containing a time value in a human-readable string format
* A named data structure (the name can be any string value), containing the bulk of the data comprising the report.

e.g. The report on MoJ Github repositories might consist of:

```
{
    "updated_at": "2020-09-16 15:23:42 UTC",
    "repositories": [ ...list of data hashes, one for each repo...]
}
```

The app. will only accept posted JSON data when the HTTP POST supplies the correct API key.

'correct' means the value of the 'X-API-KEY' header in the HTTP POST must match the value of the 'API_KEY' environment variable that was in scope when the app. was started.

If the supplied API key matches the expected value, the locally stored JSON data file will be overwritten with the request body supplied in the POST.

If the API key doesn't match, the app. will return a 403 error.

### Developing

If you have a working ruby 2.7 environment, you can run the web application locally as follows:

```
bundle install
./app.rb
```

> The cronjobs for both dev and prod are deployed into the same namespace (`manager/concourse-main`). So you need to change the chart name if you want to deploy development cronjobs alongside the existing production cronjobs.

## Updating the docker images

After code changes, create a new [release] via the github web interface.

This will trigger a github action to build all the docker images (web application and scheduled jobs), and push them to docker hub tagged with the release name.

[release]: https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/releases

---
last_reviewed_on: 2020-12-31
review_in: 3 months
---
