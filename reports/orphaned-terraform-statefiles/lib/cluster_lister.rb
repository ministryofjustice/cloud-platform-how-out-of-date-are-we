# List the short names of all the clusters which currently exist
class ClusterLister
  attr_reader :region

  def initialize(params)
    @region = params.fetch(:region)
  end

  def list
    kops_clusters + eks_clusters
  end

  private

  def kops_clusters
    json = `kops get clusters --output json`

    # If the output json has one cluster it doesnot enclose [] and hence map cannot 
    # dig into metadata. Check if has enclosed [] if not add a []
    if json.chr != '[' 
      json = "[" + json + "]"
    end

    JSON.parse(json)
      .map {|h| h.dig("metadata", "name")}
      .map {|str| str.split(".").first}
  end

  def eks_clusters
    json = `aws eks list-clusters --region=#{region} --output json`
    JSON.parse(json).fetch("clusters")
  end
end
