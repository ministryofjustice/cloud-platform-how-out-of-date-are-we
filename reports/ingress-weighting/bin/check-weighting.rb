#!/usr/bin/env ruby

# Global environments
LIVE_CLUSTER = "live"

def main
  get_kubeconfig

  print("ya")
  puts env(KUBECONFIG_AWS_ACCESS_KEY_ID)
end

def get_kubeconfig
  # How to fetch the kubeconfig file, so we can talk to the cluster
  kubeconfig = {
    s3client: Aws::S3::Client.new(
      region: env("KUBECONFIG_AWS_REGION"),
      credentials: Aws::Credentials.new(env("KUBECONFIG_AWS_ACCESS_KEY_ID"), env("KUBECONFIG_AWS_SECRET_ACCESS_KEY"))
    ),
    bucket: env("KUBECONFIG_S3_BUCKET"),
    key: env("KUBECONFIG_S3_KEY"),
    local_target: env("KUBECONFIG"),
    context: "live",
  }
  Kubeconfig.new(kubeconfig).fetch_and_store
end

