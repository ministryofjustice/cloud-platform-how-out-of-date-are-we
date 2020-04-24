#!/usr/bin/env ruby

# Output the results of `helm whatup` as JSON, with 'apps' and 'updated_at' keys.

require "json"
require "open3"

# This script expects to be invoked once the kube context for live-1 has been set up.

def get_helm_release_data
  stdout, _, _ = Open3.capture3("helm whatup -o json")
  JSON.parse(stdout)
end

clusters = [
  name: "live-1",
  apps: get_helm_release_data,
]

# Switch to the manager cluster and repeat

region = ENV.fetch("AWS_REGION")
Open3.capture3("aws eks --region #{region} update-kubeconfig --name manager --alias manager")
Open3.capture3("kubectl config use-context manager")

clusters << { name: "manager", apps: get_helm_release_data }

data = {
  clusters: clusters,
  updated_at: Time.now
}

puts data.to_json
