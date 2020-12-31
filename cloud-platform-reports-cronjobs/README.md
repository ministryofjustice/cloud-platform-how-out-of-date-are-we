# Cronjobs for Cloud Platform Reports

This helm chart defines kubernetes cronjobs which generate data and post it as
JSON to the Cloud Platform Reports web application.

These cronjobs are deployed to the `manager` cluster, in the `concourse-main`
namespace, because all the required secrets are already being maintained in
there.
