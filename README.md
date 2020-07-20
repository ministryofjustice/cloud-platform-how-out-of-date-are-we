# How out of date are we?

Simple web app. to display various status information including:

* a traffic light view of how far our installed helm charts are behind the latest versions
* documentation pages which are overdue for review.
* namespaces in the environments repository which use versions of our terraform modules which are not the latest
* ministryofjustice/cloud-platform-\* github repositories whose settings do not match our requirements

![Screenshot of the app](screenshot.png?raw=true "Example screenshot")

The app. accepts posted JSON data from an updater image, defined in the [updater-image] directory.

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

### Helm releases

To provision data to the app, make an HTTP post, like this:

    curl -H "X-API-KEY: soopersekrit" -d "$(helm whatup -o json)" http://localhost:4567/helm_whatup

JSON data should be the output of [Helm Whatup](https://github.com/bacongobbler/helm-whatup)

### Terraform Modules

To provision data to the app, make an HTTP post, like this:

    curl -H "X-API-KEY: soopersekrit" -d "[JSON data]" http://localhost:4567/terraform_modules

JSON data should be the output of the [terraform modules version checker script](updater-image/module-versions.rb)

To run the script, you need a `GITHUB_TOKEN` environment variable, containing a
GitHub personal access token which has had single sign-on (SSO) enabled for the
ministryofjustice GitHub organisation. The token does not need any scopes
enabled, since all our repos are public.

Once data has been posted, visit the app at `http://localhost:4567`

The app. will only accept posted JSON data when the HTTP POST supplies the correct API key.

'correct' means the value of the 'X-API-KEY' header in the HTTP POST must match the value of the 'API_KEY' environment variable that was in scope when the app. was started.

If the supplied API key matches the expected value, the locally stored JSON data file will be overwritten with the request body supplied in the POST.

If the API key doesn't match, the app. will return a 403 error.

### Documentation pages

This uses the `updater-image/documentation-pages-to-review.rb` script in a similar way to the terraform_modules script.

In addition to the API key, this script uses the value of the `DOCUMENTATION_SITES` environment variable to decide what sites to crawl, looking for documentation pages which are past their "review by" dates.

### Repositories

This uses: https://github.com/ministryofjustice/cloud-platform-repository-checker

It requires a github personal access token with `public_repo` scope.

### Developing

See the `docker-compose.yml` file for details of how to run this app. and the updater script locally.

If you're just working on the web application, another option is to run:

```
bundle install
make fetch-live-json-datafiles
make dev-server
```

This will launch a local instance of the web server, and populate the data
directory with the latest JSON files from the live instance.

NB: You will need a workin ruby environment.

## Updating the docker images

After code changes, create a new [release] via the github web interface.

This will trigger a github action to build both the web app. and updater docker
images, and push them to docker hub tagged with the release name.

[release]: https://github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/releases
