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
  # TODO: there is a lot of duplication here from the code for the individual docpaths. Fix that.
  updated = []

  terraform_modules, updated_at = get_list_and_updated_at(datafile("terraform_modules"), "out_of_date_modules")
  updated << updated_at

  documentation_pages, updated_at = get_list_and_updated_at(datafile("documentation"), "pages")
  updated << updated_at

  repositories, updated_at = get_list_and_updated_at(datafile("repositories"), "repositories")
  repositories.reject! { |repo| repo["status"] == "PASS" }
  updated << updated_at

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

  item_list = klass.new(
    file: datafile(docpath),
    key: key,
    logger: logger,
  )

  locals = {
    active_nav: docpath,
    updated_at: item_list.updated_at,
    list: item_list.list,
  }

  erb template, locals: locals
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
  fetch_data_and_render_template("documentation", "pages") do |list|
    list.each_with_index do |url, i|
      # Turn the URL into site/title/url tuples e.g.
      #   "https://runbooks.cloud-platform.service.justice.gov.uk/create-cluster.html" -> site: "runbooks", title: "create-cluster"
      site, _, _, _, _, title = url.split(".").map { |s| s.sub(/.*\//, '') }
      list[i] = { "site" => site, "title" => title, "url" => url }
    end
  end
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
