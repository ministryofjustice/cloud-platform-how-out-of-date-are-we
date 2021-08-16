require "bundler/setup"
require "aws-sdk-s3"
require "json"
require "kubeclient"
require "open3"

require "#{File.dirname(__FILE__)}/hosted_services/kubeconfig"
require "#{File.dirname(__FILE__)}/hosted_services/cluster_namespace_lister"
