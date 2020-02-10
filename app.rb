#!/usr/bin/env ruby

require "bundler/setup"
require "json"
require "sinatra"

if development?
  require "sinatra/reloader"
  require "pry-byebug"
end

WHATUP_JSON_FILE = "data/helm-whatup.json"

# Return success/warning/danger, depending on
# how far behind latest the installed version
# is.
def version_lag_traffic_light(app)
  installed = app.fetch("installedVersion").split(".")
  latest = app.fetch("latestVersion").split(".")

  major_diff = latest[0].to_i - installed[0].to_i
  minor_diff = latest[1].to_i - installed[1].to_i

  if major_diff > 1
    "danger"
  elsif minor_diff > 5
    "warning"
  else
    "success"
  end
end

############################################################

get "/" do
  redirect "/helm_whatup"
end

get "/helm_whatup" do
  apps = JSON.parse(File.read WHATUP_JSON_FILE)
  apps.map { |app| app["trafficLight"] = version_lag_traffic_light(app) }
  erb :helm_whatup, locals: { apps: apps }
end
