#!/usr/bin/env ruby

require "./lib/orphaned_statefiles"

s3 = Aws::S3::Resource.new(
  region: ENV.fetch("TF_STATE_BUCKET_REGION"),
  access_key_id: ENV.fetch("TF_STATE_BUCKET_AWS_ACCESS_KEY_ID"),
  secret_access_key: ENV.fetch("TF_STATE_BUCKET_AWS_SECRET_ACCESS_KEY"),
)

ctsf = DeletedClusterTerraformStateFiles.new(
  s3: s3,
  bucket: "cloud-platform-terraform-state",
  cluster_region: "eu-west-2",
)

puts({ data: ctsf.list, updated_at: Time.now }.to_json)
