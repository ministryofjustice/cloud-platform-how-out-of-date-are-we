#!/usr/bin/env ruby

require "bundler/setup"
require "json"
require "sinatra"
require "./lib/hoodaw"

CONTENT_TYPE_JSON = "application/json"

if development?
  require "sinatra/reloader"
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
    terraform_modules: get_data_from_json_file("terraform_modules", "out_of_date_modules", ItemList),
    orphaned_resources: get_data_from_json_file("orphaned_resources", "orphaned_aws_resources", OrphanedResources),
    orphaned_statefiles: get_data_from_json_file("orphaned_statefiles", "data", ItemList),
    live_1_domains: get_data_from_json_file("live_1_domains", "live_1_domains", ItemList),
  }

  updated_at = info.values.map(&:updated_at).min
  todo_count = info.values.map(&:todo_count).sum

  {
    updated_at: updated_at,
    data: {
      action_items: {
        documentation: info[:documentation].todo_count,
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
def render_item_list(params)
  docpath = params.fetch(:docpath)
  key = params.fetch(:key)
  klass = params.fetch(:klass, ItemList)
  title = params.fetch(:title, "Cloud Platform Reports")

  template = docpath.to_sym

  item_list = get_data_from_json_file(docpath, key, klass)

  locals = {
    title: title,
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
  data = costs.list.find { |ns| ns["name"] == namespace } || {}
  data.merge(
    "updated_at" => costs.updated_at,
    "resource_costs" => data["breakdown"].to_a.sort_by { |a| a[1] }.reverse
  )
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
  unless nu.namespace(namespace).nil?
    nu.namespace(namespace).merge(updated_at: nu.updated_at)
  end
end

def namespace_usage_from_json
  json = store.retrieve_file("data/namespace_usage.json")
  NamespaceUsage.new(json: json)
end

def hosted_services_for_namespace(namespace)
  json = store.retrieve_file("data/hosted_services.json")
  data = JSON.parse(json)
  ns = data["namespace_details"].find {|h| h["Name"] == namespace} || {}
  ns.merge("updated_at" => DateTime.parse(data["updated_at"]))
end

def live_1_domains_from_json
  json = store.retrieve_file("data/live_1_domains.json")
  LiveOneDomains.new(json: json)
end

def infra_deployments_from_json
  json = store.retrieve_file("data/infrastructure_deployments.json")
  InfraDeployments.new(json: json)
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

get "/documentation" do
  if accept_json?(request)
    serve_json_data(:documentation)
  else
    render_item_list(title: "Documentation", docpath: "documentation", key: "pages", klass: Documentation)
  end
end

get "/terraform_modules" do
  if accept_json?(request)
    serve_json_data(:terraform_modules)
  else
    render_item_list(title: "Terraform Modules", docpath: "terraform_modules", key: "out_of_date_modules")
  end
end

get "/orphaned_resources" do
  if accept_json?(request)
    serve_json_data(:orphaned_resources)
  else
    render_item_list(title: "Orphaned AWS Resources", docpath: "orphaned_resources", key: "orphaned_aws_resources", klass: OrphanedResources)
  end
end

get "/orphaned_statefiles" do
  if accept_json?(request)
    serve_json_data(:orphaned_statefiles)
  else
    render_item_list(title: "Orphaned Terraform Statefiles", docpath: "orphaned_statefiles", key: "data")
  end
end

get "/live_1_domains" do
  if accept_json?(request)
    serve_json_data(:live_1_domains)
  else
    lod = live_1_domains_from_json
    locals = {
      title: "Services with live-1 Domains",
      total_ingress: lod.total_ingress,
      updated_at: lod.updated_at,
      details: lod.ingress,
    }
    erb :live_1_domains, locals: locals
  end
end

get "/infrastructure_deployments" do
  if accept_json?(request)
    serve_json_data(:infrastructure_deployments)
  else
    infra = infra_deployments_from_json
    locals = {
      title: "Infrastructure deployments",
      deployments: infra.deployments,
      updated_at: infra.updated_at,
    }
    erb :infrastructure_deployments, locals: locals
  end
end


get "/costs_by_namespace" do
  if accept_json?(request)
    serve_json_data(:costs_by_namespace)
  else
    costs = namespace_costs
    locals = {
      title: "Costs by namespace",
      updated_at: costs.updated_at,
      costs: costs,
    }
    erb :costs_by_namespace, locals: locals
  end
end

get "/namespace_usage" do
  redirect "/namespace_usage_cpu"
end

get "/namespace_usage_cpu" do
  locals = namespaces_data("CPU").merge(
    column_titles: ["Namespaces", "Total pod requests (millicores)", "CPU used (millicores)"],
    title: "Namespaces by CPU (requested vs. used)",
  )
  erb :namespaces_chart, locals: locals, layout: :namespace_usage_layout
end

get "/namespace_usage_memory" do
  locals = namespaces_data("Memory").merge(
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

get "/namespace/:namespace" do
  namespace = params["namespace"]
  details = hosted_services_for_namespace(namespace)

  erb :namespace, layout: :namespace_layout, locals: {
    namespace: namespace,
    details: details,
    namespace_costs: costs_for_namespace(namespace),
    usage: usage_for_namespace(namespace),
    updated_at: details["updated_at"],
  }
end

get "/about" do
  erb :about, locals: { title: "About", updated_at: Date.parse("2020-12-30") }
end

post "/:docpath" do
  update_json_data(store, params.fetch("docpath"), request)
end
