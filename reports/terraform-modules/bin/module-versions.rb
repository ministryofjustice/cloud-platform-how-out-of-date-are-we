#!/usr/bin/env ruby

# Report any namespaces which are using old (i.e. not the latest) versions of
# any of our terraform modules

require "json"
require "net/http"
require "open3"
require "uri"

ORG = ENV.fetch("ORG")
NAMESPACE_DIR = ENV.fetch("NAMESPACE_DIR")
TF_MODULE_REGEX = ENV.fetch("TF_MODULE_REGEX")
GITHUB_API_URL = ENV.fetch("GITHUB_API_URL")
GITHUB_TOKEN = ENV.fetch("GITHUB_TOKEN")

ModuleUsage = Struct.new(:namespace, :module, :version, :latest)

def out_of_date_modules
  modules_in_use = in_use
  latest_releases = module_latest_releases(in_use)

  out_of_date_list = []

  # Compare the version in use to the latest version available
  modules_in_use.each do |module_usage|
    latest = latest_releases[module_usage.module]
    if module_usage.version != latest
      module_usage.latest = latest
      out_of_date_list << module_usage.to_h
    end
  end

  out_of_date_list
end

# Return a list of all module usage (namespace, module, version)
def in_use
  namespaces
    .sort
    .map { |ns| modules_used(ns) }
    .flatten
    .map { |line| module_usage(line) }
end

# Takes all the ModuleUsage objects that exist, reduces to unique
# module names, and returns a hash: { module name => latest version }
def module_latest_releases(modules_in_use)
  modules_in_use
    .map { |mu| mu.module }
    .uniq
    .each_with_object({}) { |mod, hash| hash[mod] = latest_version(mod); }
end

# Returns a list of namespace directory names. This relies on our
# convention of always naming the directory the same as the namespace
def namespaces
  Dir["#{NAMESPACE_DIR}/*"]
    .find_all { |dir| FileTest.directory?(dir) }
    .map { |dir| File.basename(dir) }
end

def modules_used(namespace)
  stdout, _stderr, _status = Open3.capture3("grep #{TF_MODULE_REGEX} #{tfdir(namespace)}/*")
  stdout.split("\n")
end

def module_usage(line)
  parts = line
    .sub(/"$/, "")
    .split("/")

  namespace = parts[3]
  mod, version = parts.last.split("?ref=")

  ModuleUsage.new(namespace, mod, version)
end

def tfdir(namespace)
  "#{NAMESPACE_DIR}/#{namespace}/resources"
end

# Takes a module name, returns the value of the last release defined in the
# corresponding github repo.
def latest_version(module_name)
  json = run_query(
    repo_name: File.join(ORG, module_name),
    token: GITHUB_TOKEN,
  )

  JSON.parse(json)
    .dig("data", "repository", "releases", "edges")
    .first.dig("node", "tagName")
rescue NoMethodError # we get this if we call 'dig' on nil
  # Experimental modules may not have any releases, so just return nothing
  nil
end

def run_query(params)
  repo_name = params.fetch(:repo_name)
  token = params.fetch(:token)

  json = {query: latest_release_query(repo_name)}.to_json
  headers = {"Authorization" => "bearer #{token}"}

  uri = URI.parse(GITHUB_API_URL)
  resp = Net::HTTP.post(uri, json, headers)

  resp.body
end

def latest_release_query(repo_name)
  owner, name = repo_name.sub("https://github.com/", "").split("/")

  %[
    {
      repository(owner: "#{owner}", name: "#{name}") {
        id
        releases(last: 1) {
          edges {
            node {
              id
              tagName
            }
          }
        }
      }
    }
  ]
end

############################################################

rtn = {
  updated_at: Time.now,
  out_of_date_modules: out_of_date_modules,
}

puts rtn.to_json
