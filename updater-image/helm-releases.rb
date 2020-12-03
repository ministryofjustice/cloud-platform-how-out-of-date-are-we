#!/usr/bin/env ruby

# Output the results of `helm whatup` as JSON, for each production cluster

require "json"
require "open3"

def data
  # assume the kube context for live-1 has already been set up
  clusters = [
    name: "live-1",
    apps: get_all_helm_releases,
  ]

  # Switch to the manager cluster and repeat
  region = ENV.fetch("AWS_REGION")
  execute "aws eks --region #{region} update-kubeconfig --name manager --alias manager"
  execute "kubectl config use-context manager"

  clusters << {name: "manager", apps: get_all_helm_releases}

  {
    clusters: clusters,
    updated_at: Time.now,
  }
end

def get_all_helm_releases
  namespaces_with_helm_releases.inject([]) do |list, namespace|
    list += helm_releases_in_namespace(namespace)
  end
end

def namespaces_with_helm_releases
  json = execute("helm list --all-namespaces -o json")
  data = JSON.parse(json)
  data.map { |h| h["namespace"] }.uniq.sort
end

def helm_releases_in_namespace(namespace)
  JSON.parse(execute("helm whatup -n #{namespace} -o json", true))
rescue JSON::ParserError
  []
end

def execute(cmd, allowed_to_fail = false)
  warn "Running: #{cmd}"
  stdout, stderr, status = Open3.capture3(cmd)

  unless allowed_to_fail || status.success?
    raise "Command failed: #{cmd}\n#{stderr}\n"
  end

  stdout
end

############################################################

puts data.to_json
