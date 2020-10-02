class CostsByNamespace
  def initialize(params)
    @data = JSON.parse(params.fetch(:json))
  end

  def list
    @list ||= @data
      .fetch("namespace")
      .map { |name, hash| hash["name"] = name; hash }
      .sort { |a, b| a["total"].to_f <=> b["total"].to_f }
      .reverse
  end

  def updated_at
    @data.fetch("updated_at")
  end

  def total
    list.sum { |i| i["total"].to_f }
  end
end
