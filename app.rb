#!/usr/bin/env ruby

require "bundler/setup"
require "json"
require "sinatra"
require "./helpers"
require "./lib/hoodaw"

CONTENT_TYPE_JSON = "application/json"

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

def dashboard_data
  updated = []

  item_list = get_data_from_json_file("terraform_modules", "out_of_date_modules", ItemList)
  terraform_modules = item_list.list
  updated << item_list.updated_at

  item_list = get_data_from_json_file("documentation", "pages", Documentation)
  documentation_pages = item_list.list
  updated << item_list.updated_at

  item_list = get_data_from_json_file("repositories", "repositories", GithubRepositories)
  repositories = item_list.list
  updated << item_list.updated_at

  clusters, updated_at = get_list_and_updated_at(datafile("helm_whatup"), "clusters")
  out_of_date_apps = clusters.map { |cluster| cluster.fetch("apps") }.flatten
    .filter { |app| version_lag_traffic_light(app) == "danger" }
  updated << updated_at

  {
    updated_at: updated.compact.sort.first,
    data: {
      action_items: {
        helm_whatup: out_of_date_apps.length,
        terraform_modules: terraform_modules.length,
        documentation: documentation_pages.length,
        repositories: repositories.length,
      },
      action_required: true
    }
  }
end

# key is the name of the key in our datafile which contains the list of
# elements we're interested in.
def render_item_list(docpath, key, klass = ItemList)
  template = docpath.to_sym

  item_list = get_data_from_json_file(docpath, key, klass)

  locals = {
    active_nav: docpath,
    updated_at: item_list.updated_at,
    list: item_list.list,
  }

  erb template, locals: locals
end

def get_data_from_json_file(docpath, key, klass)
  klass.new(
    file: datafile(docpath),
    key: key,
    logger: logger,
  )
end

# key is the name of the key in our datafile which contains the list of
# elements we're interested in.
def fetch_data_and_render_template(docpath, key)
  file = datafile(docpath)
  template = docpath.to_sym
  locals = {
    active_nav: docpath,
    updated_at: nil,
    list: []
  }

  if FileTest.exists?(file)
    list, updated_at = get_list_and_updated_at(file, key)

    # Do any pre-processing to the list we get from the data file
    yield list if block_given?

    locals.merge!(
      updated_at: updated_at,
      list: list
    )
  end

  erb template, locals: locals
end

def get_list_and_updated_at(file, key)
  list = []
  updated_at = nil

  begin
    data = JSON.parse(File.read file)
    updated_at = string_to_formatted_time(data.fetch("updated_at"))
    list = data.fetch(key)
  rescue JSON::ParserError
    logger.info "Malformed JSON file: #{file}"
  end

  [list, updated_at]
end

get "/" do
  redirect "/dashboard"
end

get "/dashboard" do
  accept = request.env["HTTP_ACCEPT"]

  if accept == CONTENT_TYPE_JSON
    dashboard_data.to_json
  else
    locals = dashboard_data.merge(
      active_nav: "dashboard",
    )
    erb :dashboard, locals: locals
  end
end

get "/helm_whatup" do
  fetch_data_and_render_template("helm_whatup", "clusters") do |clusters|
    clusters.each do |cluster|
      cluster.fetch("apps").map { |app| app["traffic_light"] = version_lag_traffic_light(app) }
    end
  end
end

get "/documentation" do
  render_item_list("documentation", "pages", Documentation)
end

get "/terraform_modules" do
  render_item_list("terraform_modules", "out_of_date_modules")
end

get "/repositories" do
  render_item_list("repositories", "repositories", GithubRepositories)
end

post "/:docpath" do
  update_json_datafile(params.fetch("docpath"), request)
end
