#!/usr/bin/env ruby

require "bundler/setup"
require "json"
require "sinatra"

if development?
  require "sinatra/reloader"
  require "pry-byebug"
end

WHATUP_JSON_FILE = "./data/helm-whatup.json"
TF_MODULES_JSON_FILE = "./data/module-versions.json"

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
  elsif minor_diff > 4
    "warning"
  else
    "success"
  end
end

def correct_api_key?(request)
  expected_key = ENV.fetch("API_KEY")
  provided_key = request.env.fetch("HTTP_X_API_KEY", "dontsetthisvalueastheapikey")

  expected_key == provided_key
end

############################################################

get "/" do
  redirect "/helm_whatup"
end

get "/helm_whatup" do
  data = JSON.parse(File.read WHATUP_JSON_FILE)
  apps = data.fetch("apps")
  updated_at = data.fetch("updated_at")
  apps.map { |app| app["trafficLight"] = version_lag_traffic_light(app) }
  erb :helm_whatup, locals: {
    active_nav: "helm_whatup",
    apps: apps,
    updated_at: updated_at
  }
end

get "/terraform_modules" do
  modules = []
  updated_at = ""

  if FileTest.exists?(TF_MODULES_JSON_FILE)
    data = JSON.parse(File.read TF_MODULES_JSON_FILE)
    updated_at = data.fetch("updated_at")
    modules = data.fetch("out_of_date_modules")
  end

  erb :terraform_modules, locals: {
    active_nav: "terraform_modules",
    modules: modules,
    updated_at: updated_at
  }
end

post "/update-data" do
  if correct_api_key?(request)
    payload = request.body.read
    data = {
      "apps" => JSON.parse(payload),
      "updated_at" => Time.now.strftime("%Y-%m-%d %H:%M:%S")
    }
    File.open(WHATUP_JSON_FILE, "w") {|f| f.puts(data.to_json)}
    status 200
  else
    status 403
  end
end
