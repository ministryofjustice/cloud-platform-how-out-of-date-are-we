#!/usr/bin/env ruby

require "bundler/setup"
require "json"
require "sinatra"

if development?
  require "sinatra/reloader"
  require "pry-byebug"
end

WHATUP_JSON_FILE = "data/helm-whatup.json"

############################################################

get "/" do
  redirect "/helm_whatup"
end

get "/helm_whatup" do
  apps = JSON.parse(File.read WHATUP_JSON_FILE)
  erb :helm_whatup, locals: { apps: apps }
end
