#!/usr/bin/env ruby

require "aws-sdk-costexplorer"
require "date"
require "json"
require "open3"
require "tempfile"

require_relative "./lib/aws_costs_by_namespace"

def main
  json = AwsCostsByNamespace.new(date: Date.today).report.to_json
  post_json(json)
end

def post_json(json)
  file = Tempfile.new("aws_costs_by_namespace")
  begin
    file.write(json)
    api_key = ENV.fetch("HOODAW_API_KEY")
    host = ENV.fetch("HOODAW_HOST")
    url = "#{host}/costs_by_namespace"
    cmd = %[curl -H "X-API-KEY: #{api_key}" -d "$(cat #{file.path})" #{url}]
    execute cmd
  ensure
    file.close
    file.unlink # deletes the temp file
  end
end

def execute(cmd)
  # puts "CMD: #{cmd}"
  stdout, stderr, status = Open3.capture3(cmd)
  # puts "OUTPUT:\n#{stdout}"
  unless status.success?
    puts "ERROR: #{stderr}"
    exit 1
  end
  stdout
end

main
