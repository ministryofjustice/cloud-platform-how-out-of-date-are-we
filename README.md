# How out of date are we?

Simple web app. to display a traffic light view of how far our installed software is behind the current versions.

Initially, this will consume JSON created by [Helm Whatup](https://github.com/bacongobbler/helm-whatup)

## Updating the JSON data

To provision data to the app, make an HTTP post, like this:

    curl -d "$(helm whatup -o json)" http://localhost:4567/update-data

Once data has been posted, visit the app at `http://localhost:4567`

