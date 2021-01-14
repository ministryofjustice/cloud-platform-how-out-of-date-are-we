module OrphanedResources
  class Reporter
    def run
      s3 = Aws::S3::Resource.new(region: "eu-west-1", profile: ENV["AWS_PROFILE"])
      ec2 = Aws::EC2::Client.new(region: "eu-west-2", profile: ENV["AWS_PROFILE"])
      route53 = Aws::Route53::Client.new(region: "eu-west-2", profile: ENV["AWS_PROFILE"])
      rds = Aws::RDS::Client.new(region: "eu-west-2", profile: ENV["AWS_PROFILE"])

      @aws = OrphanedResources::AwsResources.new(
        s3client: s3,
        ec2client: ec2,
        route53client: route53,
        rdsclient: rds,
      )

      @terraform = OrphanedResources::TerraformStateManager.new(
        s3client: s3,
        bucket: "cloud-platform-terraform-state",
        cache_dir: "state-files"
      )

      {
        nat_gateways: compare(:nat_gateways),
        hosted_zones: compare(:hosted_zones),
        internet_gateways: compare(:internet_gateways),
        subnets: compare(:subnets),
        vpcs: compare(:vpcs),
        route_tables: compare(:route_tables),
        route_table_associations: compare(:route_table_associations),
        rds: compare(:rds),
        rds_cluster: compare(:rds_cluster),
        kops_cluster: orphaned_kops_clusters,
      }
    end

    private

    def compare(method)
      ResourceTuple.subtract_lists(
        @aws.send(method),
        @terraform.send(method)
      ).sort
    end

    def orphaned_kops_clusters
      a = @aws.kops_clusters
      t = @terraform.kops_clusters
      orphaned = (a.map { |i| i[:cluster]}) - (t.map { |i| i[:cluster]})
      a.filter { |c| orphaned.include?(c[:cluster]) }
    end
  end
end
