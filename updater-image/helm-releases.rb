#!/usr/bin/env ruby

# Output the results of `helm whatup` as JSON, with 'apps' and 'updated_at' keys.

require "json"
require "open3"

stdout, _, _ = Open3.capture3("helm whatup -o json")
apps = JSON.parse(stdout)
data = {
  clusters: [
    name: "live-1",
    apps: apps,
  ],
  updated_at: Time.now
}
puts data.to_json
