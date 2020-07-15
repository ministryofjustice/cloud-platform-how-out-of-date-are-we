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

def require_api_key(request)
  if correct_api_key?(request)
    yield
    status 200
  else
    status 403
  end
end

def datafile(docpath)
  "./data/#{docpath}.json"
end

def dashboard_data
  info = {
    documentation: get_data_from_json_file("documentation", "pages", Documentation),
    helm_whatup: get_data_from_json_file("helm_whatup", "clusters", HelmWhatup),
    repositories: get_data_from_json_file("repositories", "repositories", GithubRepositories),
    terraform_modules: get_data_from_json_file("terraform_modules", "out_of_date_modules", ItemList),
  }

  updated_at = info.values.map(&:updated_at).sort.first
  todo_count = info.values.map(&:todo_count).sum

  {
    updated_at: updated_at,
    data: {
      action_items: {
        documentation: info[:documentation].todo_count,
        helm_whatup: info[:helm_whatup].todo_count,
        repositories: info[:repositories].todo_count,
        terraform_modules: info[:terraform_modules].todo_count,
      },
      action_required: (todo_count > 0),
    }
  }
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

############################################################

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
  render_item_list("helm_whatup", "clusters", HelmWhatup)
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
