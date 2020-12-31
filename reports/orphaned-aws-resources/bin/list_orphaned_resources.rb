#!/usr/bin/env ruby

# List all resources on AWS which are not mentioned in any terraform state
# files.

require_relative "../lib/orphaned_resources"

report = {
  orphaned_aws_resources: OrphanedResources::Reporter.new.run,
  updated_at: Time.now
}

puts report.to_json
