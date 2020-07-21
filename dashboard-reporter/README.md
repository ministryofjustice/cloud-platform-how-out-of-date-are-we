# Dashboard Reporter

This directory maintains a `report.rb` ruby script which
consumes the HOODAW dashboard JSON endpoint and posts a summary
to the team slack channel iff there are outstanding action
items.

The main `DashboardReporter` class is defined in the `lib`
directory (a sibling to this directory) so that we can take
advantage of the rspec test framework.

The `makefile` in this directory creates `report.rb` by
concatenating the `../lib/dashboard_reporter.rb` file with a
header and `footer.rb` file.

The `.github/workflows/docker-hub.yml` file defines a github
action with runs the makefile and then builds and pushes the
docker image to docker hub whenever a new release of this
project is created.
