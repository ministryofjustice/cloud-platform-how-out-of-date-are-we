#!/usr/bin/env ruby

require "bundler/setup"
require "json"
require "sinatra"
require "./lib/hoodaw"

CONTENT_TYPE_JSON = "application/json"

if development?
  require "sinatra/reloader"
  require "pry-byebug"
end

def update_json_datafile(docpath, request)
  require_api_key(request) do
    file = datafile(docpath)
    dir = File.dirname(file)

    FileUtils.mkdir_p(dir) unless FileTest.directory?(dir)
    File.write(file, request.body.read)
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

def correct_api_key?(request)
  expected_key = ENV.fetch("API_KEY")
  provided_key = request.env.fetch("HTTP_X_API_KEY", "dontsetthisvalueastheapikey")

  expected_key == provided_key
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
    orphaned_resources: get_data_from_json_file("orphaned_resources", "orphaned_aws_resources", OrphanedResources),
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
        orphaned_resources: info[:orphaned_resources].todo_count,
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

def serve_json_data(docpath)
  File.read(datafile(docpath))
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

def accept_json?(request)
  accept = request.env["HTTP_ACCEPT"]
  accept == CONTENT_TYPE_JSON
end

############################################################

get "/" do
  redirect "/dashboard"
end

get "/dashboard" do
  if accept_json?(request)
    dashboard_data.to_json
  else
    locals = dashboard_data.merge(
      active_nav: "dashboard",
    )
    erb :dashboard, locals: locals
  end
end

get "/helm_whatup" do
  if accept_json?(request)
     serve_json_data(:helm_whatup)
  else
    render_item_list("helm_whatup", "clusters", HelmWhatup)
  end
end

get "/documentation" do
  if accept_json?(request)
     serve_json_data(:documentation)
  else
    render_item_list("documentation", "pages", Documentation)
  end
end

get "/terraform_modules" do
  if accept_json?(request)
     serve_json_data(:terraform_modules)
  else
    render_item_list("terraform_modules", "out_of_date_modules")
  end
end

get "/repositories" do
  if accept_json?(request)
     serve_json_data(:repositories)
  else
    render_item_list("repositories", "repositories", GithubRepositories)
  end
end

get "/orphaned_resources" do
  if accept_json?(request)
     serve_json_data(:orphaned_resources)
  else
    render_item_list("orphaned_resources", "orphaned_aws_resources", OrphanedResources)
  end
end

get "/namespace_costs" do
  if accept_json?(request)
     # TODO: figure out what to do here
  else
    nc = NamespaceCosts.new(dir: "data/namespace/costs")
    locals = {
      active_nav: "namespace_costs",
      updated_at: nc.updated_at,
      list: nc.list,
      total: nc.total,
    }
    erb :namespace_costs, locals: locals
  end
end

get "/namespace_cost/:namespace" do
  if accept_json?(request)
     # TODO: figure out what to do here
  else
    namespace_cost = NamespaceCost.new(file: "data/namespace/costs/#{params.fetch("namespace")}.json")
    locals = {
      active_nav: "namespace_costs",
      namespace_cost: namespace_cost,
      updated_at: namespace_cost.updated_at,
    }
    erb :namespace_cost, locals: locals
  end
end

post "/:docpath" do
  update_json_datafile(params.fetch("docpath"), request)
end

post "/namespace/costs/:namespace" do
  path = "namespace/costs/#{params.fetch("namespace")}"
  update_json_datafile(path, request)
end
