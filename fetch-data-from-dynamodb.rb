#!/usr/bin/env ruby

# Monkey-patch Object so that we can require lib/hoodaw.rb without requiring
# sinatra
class Object
  def helpers(_)
  end
end

require "bundler/setup"
require "pry-byebug"
require "json"
require "./lib/hoodaw"

store = Dynamodb.new
files = store.list_files
store.retrieve_files(files).each { |file, data| puts file; File.write(file, data["content"]) }
