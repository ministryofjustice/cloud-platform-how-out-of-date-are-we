module OrphanedResources
  # Class to represent an AWS resource, for the purposes of the orphaned resources report
  class ResourceTuple
    attr_reader :id, :cluster, :aws_console_url

    def initialize(params)
      @id = params.fetch(:id)
      @cluster = params.fetch(:cluster, "")
      @aws_console_url = params.fetch(:aws_console_url, "")
    end

    # Subtract one list of ResourceTuple objects from another
    def self.subtract_lists(orig, subtract)
      ids = subtract.map(&:id)
      orig.reject { |tuple| ids.include?(tuple.id) }
    end

    def add_cluster_tag(resource)
      if resource.respond_to?(:tags)
        t = resource.tags.find { |tag| tag.key.downcase == "cluster" }
        @cluster = t.value unless t.nil?
      end
      self
    end

    # Enable `.compact` to work on lists of these
    def empty?
      id.nil?
    end

    # Enable `.sort` to work on lists of these
    def <=>(other)
      id <=> other.id
    end

    # Enable `.to_json` to work on lists of these
    def to_json(_json_ext_generator_state)
      {
        id: id,
        cluster: cluster,
        aws_console_url: aws_console_url,
      }.to_json
    end
  end
end
