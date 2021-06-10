module OrphanedResources
  class TerraformStateManager < Lister
    attr_reader :s3client, :bucket, :cache_dir

    CLUSTER_SUFFIX = "cloud-platform.service.justice.gov.uk"

    def initialize(args)
      @s3client = args.fetch(:s3client)
      @bucket = args.fetch(:bucket)
      @cache_dir = args.fetch(:cache_dir)
    end

    def local_statefiles
      @files ||= download_files
    end

    def kops_clusters
      cluster_names = local_statefiles
        .grep(/^state-files\/aws-accounts\/cloud-platform-aws\/vpc\/kops\//)
        .map {|f| f.split("/")[5] }  # e.g. "live-1", "cp-0401-1622"
        .map { |c| [c, CLUSTER_SUFFIX].join(".") } # e.g. "live-1.cloud-platform.service.justice.gov.uk"

      # Return a list of tuples with the same structure as the one we get from
      # AwsResources#kops_clusters
      cluster_names.map { |c| { cluster: c, instances: nil } }
    end

    def vpcs
      list = local_statefiles.map { |file|
        data = parse_json(file)
        data.dig("outputs", "network_id", "value")
      }
      clean_list(list).map { |id| ResourceTuple.new(id: id) }
    end

    def nat_gateways
      list = local_statefiles.inject([]) { |ids, file| ids << nat_gateway_ids_from_statefile(file) }
      clean_list(list).map { |id| ResourceTuple.new(id: id) }
    end

    def subnets
      list = local_statefiles.inject([]) { |ids, file| ids << subnet_ids_from_statefile(file) }
      clean_list(list).map { |id| ResourceTuple.new(id: id) }
    end

    def route_tables
      list = local_statefiles.inject([]) { |ids, file| ids << route_tables_from_statefile(file) }
      clean_list(list).map { |id| ResourceTuple.new(id: id) }
    end

    def route_table_associations
      list = local_statefiles.inject([]) { |ids, file| ids << route_table_associations_from_statefile(file) }
      clean_list(list).map { |id| ResourceTuple.new(id: id) }
    end

    def internet_gateways
      list = local_statefiles.inject([]) { |ids, file| ids << internet_gateways_from_statefile(file) }
      clean_list(list).map { |id| ResourceTuple.new(id: id) }
    end

    def hosted_zones
      list = local_statefiles.inject([]) { |ids, file| ids << hosted_zones_from_statefile(file) }
      clean_list(list).map { |id| ResourceTuple.new(id: id) }
    end

    def rds
      list = local_statefiles.inject([]) { |ids, file| ids << rds_from_statefile(file) }
      clean_list(list).map { |id| ResourceTuple.new(id: id) }
    end

    def rds_cluster
      list = local_statefiles.inject([]) { |ids, file| ids << rds_cluster_from_statefile(file) }
      clean_list(list).map { |id| ResourceTuple.new(id: id) }
    end

    private

    def internet_gateways_from_statefile(file)
      json_resources(file)
        .filter {|h| h["type"] == "aws_internet_gateway"}
        .map { |h| h["instances"]}
        .flatten
        .map {|h| h.dig("attributes", "id")}
    end

    def route_tables_from_statefile(file)
      data = parse_json(file)
      data.dig("outputs", "private_route_tables", "value").to_a + data.dig("outputs", "public_route_tables", "value").to_a
    end

    def route_table_associations_from_statefile(file)
      json_resources(file)
        .find_all { |res| res["type"] == "aws_route_table_association" }
        .map { |res| res["instances"] }
        .flatten
        .map { |res| res.dig("attributes", "id") }
    end

    def subnet_ids_from_statefile(file)
      data = parse_json(file)

      resource_instance_values = data.fetch("resources", []).map {|h| h.fetch("instances", [])}.flatten.map { |h| h.dig("attributes", "outputs", "value") }.compact

      external = resource_instance_values.map { |h| h["external_subnets_ids"] }.compact
      internal = resource_instance_values.map { |h| h["internal_subnets_ids"] }.compact

      (external + internal).sort.uniq
    end

    def nat_gateway_ids_from_statefile(file)
      json_resources(file)
        .find_all { |hash| hash["name"] = "private_nat_gateway" }
        .map { |hash| hash["instances"] }
        .flatten
        .map { |hash| hash.dig("attributes", "nat_gateway_id") }
        .compact
    end

    def hosted_zones_from_statefile(file)
      json_resources(file)
        .find_all { |res| res["type"] == "aws_route53_zone" }
        .map { |zone| zone["instances"] }
        .flatten
        .map { |inst| inst.dig("attributes", "name") }
        .map { |name| name.sub(/\.$/, "") } # trim trailing '.'
    end

    def rds_from_statefile(file)
      json_resources(file)
      .find_all { |r| r["type"] == "aws_db_instance" || r["type"] == "aws_rds_cluster_instance" }
      .map { |rds| rds["instances"] }.flatten
      .map { |i| i.dig("attributes", "identifier") }
    end

    def rds_cluster_from_statefile(file)
      json_resources(file)
      .find_all { |r| r["type"] == "aws_rds_cluster" }
      .map { |rds| rds["instances"] }.flatten
      .map { |i| i.dig("attributes", "cluster_identifier") }
    end

    def download_files
      s3client.bucket(bucket)
        .objects
        .collect(&:key)
        .find_all { |key| key =~ /terraform.tfstate$/ }
        .map { |key| download_file(key) }
    end

    def download_file(key)
      outfile = File.join(cache_dir, key)
      d = File.dirname(outfile)
      FileUtils.mkdir_p(d) unless Dir.exist?(d)
      s3client.bucket(bucket).object(key).get(response_target: outfile) unless FileTest.exists?(outfile)
      outfile
    end

    def json_resources(file)
      data = parse_json(file)
      data.fetch("resources", [])
    end

    def parse_json(file)
      JSON.parse(File.read(file))
    rescue JSON::ParserError
      {}
    end
  end
end
