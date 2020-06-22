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

# key is the name of the key in our datafile which contains the list of
# elements we're interested in.
def fetch_data(docpath, key)
  file = datafile(docpath)
  template = docpath.to_sym
  locals = {
    active_nav: docpath,
    updated_at: nil,
    list: []
  }

  if FileTest.exists?(file)
    data = {}
    list = []
    updated_at = nil

    begin
      data = JSON.parse(File.read file)
      updated_at = string_to_formatted_time(data.fetch("updated_at"))
      list = data.fetch(key)
    rescue JSON::ParserError
      logger.info "Malformed JSON file: #{file}"
    end

    # Do any pre-processing to the list we get from the data file
    yield list if block_given?

    locals.merge!(
      updated_at: updated_at,
      list: list
    )
  end

  erb template, locals: locals
end

get "/" do
  redirect "/dashboard"
end

get "/dashboard" do
  locals = {
    active_nav: "/dashboard",
    updated_at: nil,
    data: {
      action_items: {
        helm_whatup: 1,
        terraform_modules: 2,
        documentation: 0,
        repositories: 3
      },
      action_required: true
    }
  }
  erb :dashboard, locals: locals
end

get "/helm_whatup" do
  fetch_data("helm_whatup", "clusters") do |clusters|
    clusters.each do |cluster|
      cluster.fetch("apps").map { |app| app["traffic_light"] = version_lag_traffic_light(app) }
    end
  end
end

get "/documentation" do
  fetch_data("documentation", "pages") do |list|
    list.each_with_index do |url, i|
      # Turn the URL into site/title/url tuples e.g.
      #   "https://runbooks.cloud-platform.service.justice.gov.uk/create-cluster.html" -> site: "runbooks", title: "create-cluster"
      site, _, _, _, _, title = url.split(".").map { |s| s.sub(/.*\//, '') }
      list[i] = { "site" => site, "title" => title, "url" => url }
    end
  end
end

get "/terraform_modules" do
  fetch_data("terraform_modules", "out_of_date_modules")
end

get "/repositories" do
  fetch_data("repositories", "repositories") do |list|
    list.reject! { |repo| repo["status"] == "PASS" }
  end
end

post "/:docpath" do
  update_json_datafile(params.fetch("docpath"), request)
end
