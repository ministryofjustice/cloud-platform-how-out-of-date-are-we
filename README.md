# How out of date are we?

Simple web app. to display a traffic light view of how far our installed software is behind the current versions.

![Screenshot of the app](screenshot.png?raw=true "Example screenshot")

Initially, this will consume JSON created by [Helm Whatup](https://github.com/bacongobbler/helm-whatup)

## Updating the JSON data

### Helm releases

To provision data to the app, make an HTTP post, like this:

    curl -H "X-API-KEY: soopersekrit" -d "$(helm whatup -o json)" http://localhost:4567/helm_whatup

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

### Developing

See the `docker-compose.yml` file for details of how to run this app. and the updater script locally.

### Updater image

The `updater-image/` directory maintains a docker image which can be used to update the JSON data in the app. See the `makefile` in that directory for a usage example.

## Updating the docker images

Pre-requisites: You need push access to the `ministryofjustice` repo on [docker hub]

To update the app. docker image:

 * make and commit your changes
 * update the tag value of `IMAGE` in the `makefile`
 * run `make`

To update the updater image, repeat these steps in the `updater-image/` directory.

This will build the docker images and push them to docker hub, using the updated tag values.
