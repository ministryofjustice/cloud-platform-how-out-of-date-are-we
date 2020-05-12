#!/usr/bin/env ruby

require "bundler/setup"
require "json"
require "sinatra"
require "./helpers"

if development?
  require "sinatra/reloader"
  require "pry-byebug"
end

def update_json_datafile(docpath, request)
  require_api_key(request) do
    File.open(datafile(docpath), "w") {|f| f.puts(request.body.read)}
  end
end

def datafile(docpath)
  "./data/#{docpath}.json"
end

def require_api_key(request)
  if correct_api_key?(request)
    yield
    status 200
  else
    status 403
  end
end

get "/" do
  redirect "/helm_whatup"
end

get "/helm_whatup" do
  clusters = []
  updated_at = nil

  file = datafile("helm_whatup")
  if FileTest.exists?(file)
    data = JSON.parse(File.read file)
    updated_at = string_to_formatted_time(data.fetch("updated_at"))
    clusters = data.fetch("clusters")
    clusters.each do |cluster|
      cluster.fetch("apps").map { |app| app["trafficLight"] = version_lag_traffic_light(app) }
    end
  end

  erb :helm_whatup, locals: {
    active_nav: "helm_whatup",
    clusters: clusters,
    updated_at: updated_at
  }
end

get "/terraform_modules" do
  modules = []
  updated_at = ""

  file = datafile("terraform_modules")
  if FileTest.exists?(file)
    data = JSON.parse(File.read file)
    updated_at = string_to_formatted_time(data.fetch("updated_at"))
    modules = data.fetch("out_of_date_modules")
  end

  erb :terraform_modules, locals: {
    active_nav: "terraform_modules",
    modules: modules,
    updated_at: updated_at
  }
end

get "/documentation" do
  pages = []
  updated_at = ""

  file = datafile("documentation")
  if FileTest.exists?(file)
    data = JSON.parse(File.read file)
    updated_at = string_to_formatted_time(data.fetch("updated_at"))
    pages = data.fetch("pages").inject([]) do |arr, url|
      # Turn the URL into site/title/url tuples e.g.
      #   "https://runbooks.cloud-platform.service.justice.gov.uk/create-cluster.html" -> site: "runbooks", title: "create-cluster"
      site, _, _, _, _, title = url.split(".").map { |s| s.sub(/.*\//, '') }
      arr << { "site" => site, "title" => title, "url" => url }
    end
  end

  erb :documentation, locals: {
    active_nav: "documentation",
    pages: pages,
    updated_at: updated_at
  }
end

post "/:docpath" do
  update_json_datafile(params.fetch("docpath"), request)
end
