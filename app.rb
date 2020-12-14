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

def update_json_data(store, docpath, request)
  require_api_key(request) do
    file = datafile(docpath)
    store.store_file(file, request.body.read)
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
  "data/#{docpath}.json"
end

def dashboard_data
  info = {
    documentation: get_data_from_json_file("documentation", "pages", Documentation),
    helm_whatup: get_data_from_json_file("helm_whatup", "clusters", HelmWhatup),
    repositories: get_data_from_json_file("repositories", "repositories", GithubRepositories),
    terraform_modules: get_data_from_json_file("terraform_modules", "out_of_date_modules", ItemList),
    orphaned_resources: get_data_from_json_file("orphaned_resources", "orphaned_aws_resources", OrphanedResources),
    orphaned_statefiles: get_data_from_json_file("orphaned_statefiles", "data", ItemList),
    hosted_services: get_data_from_json_file("hosted_services", "namespace_details", ItemList),
  }

  updated_at = info.values.map(&:updated_at).min
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
        orphaned_statefiles: info[:orphaned_statefiles].todo_count,
      },
      action_required: (todo_count > 0),
    },
  }
end

def get_data_from_json_file(docpath, key, klass)
  klass.new(
    store: store,
    file: datafile(docpath),
    key: key,
    logger: logger,
  )
end

def serve_json_data(docpath)
  store.retrieve_file(datafile(docpath))
end

# key is the name of the key in our datafile which contains the list of
# elements we're interested in.
def render_item_list(docpath, key, klass = ItemList)
  template = docpath.to_sym

  item_list = get_data_from_json_file(docpath, key, klass)

  locals = {
    updated_at: item_list.updated_at,
    list: item_list.list,
  }

  erb template, locals: locals
end

def accept_json?(request)
  accept = request.env["HTTP_ACCEPT"]
  accept == CONTENT_TYPE_JSON
end

def store
  ENV.key?("DYNAMODB_TABLE_NAME") ? Dynamodb.new : Filestore.new
end

def costs_for_namespace(namespace)
  costs = namespace_costs
  data = costs.list.find { |ns| ns["name"] == namespace }
  data.merge("updated_at" => costs.updated_at)
end

def namespace_costs
  json = store.retrieve_file datafile("costs_by_namespace")
  CostsByNamespace.new(json: json)
end

def namespaces_data(order_by)
  nu = namespace_usage_from_json

  {
    values: nu.values(order_by),
    updated_at: nu.updated_at,
    type: order_by,
    total_requested: nu.total_requested(order_by), # order_by is cpu|memory
  }
end

def usage_for_namespace(namespace)
  nu = namespace_usage_from_json
  nu.namespace(namespace).merge(updated_at: nu.updated_at)
end

def namespace_usage_from_json
  json = store.retrieve_file("data/namespace_usage.json")
  NamespaceUsage.new(json: json)
end

def hosted_services_for_namespace(namespace)
  data = JSON.parse File.read(datafile("hosted_services"))
  data["namespace_details"].find {|h| h["namespace"] == namespace}
end

############################################################

get "/" do
  redirect "/dashboard"
end

get "/dashboard" do
  if accept_json?(request)
    dashboard_data.to_json
  else
    erb :dashboard, locals: dashboard_data
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

get "/orphaned_statefiles" do
  if accept_json?(request)
    serve_json_data(:orphaned_statefiles)
  else
    render_item_list("orphaned_statefiles", "data", ItemList)
  end
end

get "/hosted_services" do
  if accept_json?(request)
    serve_json_data(:hosted_services)
  else
    render_item_list("hosted_services", "namespace_details")
  end
end

get "/costs_by_namespace" do
  costs = namespace_costs

  locals = {
    updated_at: costs.updated_at,
    costs: costs,
  }

  erb :costs_by_namespace, locals: locals
end

get "/namespace_cost/:namespace" do
  namespace_cost = costs_for_namespace(params["namespace"])

  if accept_json?(request)
    namespace_cost.to_json
  else
    # Sort costs in reverse value order
    resource_costs = namespace_cost["breakdown"].to_a.sort_by { |a| a[1] }.reverse

    locals = {
      namespace: namespace_cost["name"],
      total: namespace_cost["total"],
      resource_costs: resource_costs,
      updated_at: namespace_cost["updated_at"],
    }
    erb :namespace_cost, locals: locals
  end
end

get "/namespace_usage" do
  redirect "/namespace_usage_cpu"
end

get "/namespace_usage_cpu" do
  locals = namespaces_data("cpu").merge(
    column_titles: ["Namespaces", "Total pod requests (millicores)", "CPU used (millicores)"],
    title: "Namespaces by CPU (requested vs. used)",
  )
  erb :namespaces_chart, locals: locals, layout: :namespace_usage_layout
end

get "/namespace_usage_memory" do
  locals = namespaces_data("memory").merge(
    column_titles: ["Namespaces", "Total pods requests (mebibytes)", "Memory used (mebibytes)"],
    title: "Namespaces by Memory (requested vs. used)",
  )
  erb :namespaces_chart, locals: locals, layout: :namespace_usage_layout
end

get "/namespace_usage_pods" do
  nu = namespace_usage_from_json

  locals = {
    column_titles: ["Namespaces", "Pods limit", "Pods running"],
    title: "Namespaces by pods (limit vs. running)",
    values: nu.pods_values,
    updated_at: nu.updated_at,
    type: "pods",
    total_requested: nu.total_pods,
  }
  erb :namespaces_chart, locals: locals, layout: :namespace_usage_layout
end

get "/namespace_usage/:namespace" do
  nu = usage_for_namespace(params[:namespace])

  erb :namespace_usage, locals: {
    data: nu,
    updated_at: nu[:updated_at],
  }
end

get "/namespace/:namespace" do
  namespace_cost = costs_for_namespace(params["namespace"])
  # Sort costs in reverse value order
  resource_costs = namespace_cost["breakdown"].to_a.sort_by { |a| a[1] }.reverse
  # TODO DRY up

  erb :namespace, layout: :namespace_layout, locals: {
    namespace: params[:namespace],
    updated_at: Time.now, # TODO fix this
    details: hosted_services_for_namespace(params["namespace"]),
    resource_costs: resource_costs,
    usage: usage_for_namespace(params[:namespace]),
    total: namespace_cost["total"], # TODO tidy up
  }
end

post "/:docpath" do
  update_json_data(store, params.fetch("docpath"), request)
end
