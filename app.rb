#!/usr/bin/env ruby

require "bundler/setup"
require "json"
require "sinatra"

if development?
  require "sinatra/reloader"
  require "pry-byebug"
end

############################################################

get "/" do
  redirect "/helm_whatup"
end

get "/helm_whatup" do
  erb :helm_whatup
end

