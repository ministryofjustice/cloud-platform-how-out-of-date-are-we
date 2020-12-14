RSpec.configure do |config|
  config.expect_with :rspec do |expectations|
    expectations.include_chain_clauses_in_custom_matcher_descriptions = true
  end

  config.mock_with :rspec do |mocks|
    mocks.verify_partial_doubles = true
  end

  config.shared_context_metadata_behavior = :apply_to_host_groups
end

require "net/http"
require "uri"
require "json"
require "sinatra"
require "./lib/hoodaw"
require "./lib/dashboard_reporter"

def fetch_url(url, accept = nil)
  uri = URI.parse(url)
  req = Net::HTTP::Get.new(uri)
  unless accept.nil?
    req["Accept"] = accept
  end
  Net::HTTP.start(uri.hostname, uri.port) { |http|
    http.request(req)
  }
end

def post_to_url(url, body, api_key = nil)
  uri = URI.parse(url)
  http = Net::HTTP.new(uri.host, uri.port)
  request = Net::HTTP::Post.new(uri.request_uri)
  request["X-API-KEY"] = api_key unless api_key.nil?
  request.body = body
  http.request(request)
end
