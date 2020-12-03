class NamespaceUsage
  def initialize(params)
    @data = JSON.parse(params.fetch(:json))
  end

  def values(order_by)
    namespaces
      .map { |n| namespace_values(n, order_by) }
      .sort_by { |i| i[1] }
      .reverse
  end

  def pods_values
    namespaces
      .map { |n| namespace_pods_values(n) }
      .sort_by { |i| i[1] }
      .reverse
  end

  def total_pods
    namespaces.sum { |n| n.dig("resources_used", "pods") }
  end

  def updated_at
    DateTime.parse(@data.fetch("updated_at"))
  end

  def total_requested(property)
    namespaces.map { |ns| ns.dig("resources_requested", property) }.map(&:to_i).sum
  end

  def namespace(name)
    namespaces.find { |n| n["name"] == name }
  end

  private

  def namespace_pods_values(namespace)
    [
      namespace.fetch("name").to_s,
      namespace.dig("hard_limit", "pods").to_i,
      namespace.dig("resources_used", "pods").to_i,
    ]
  end

  def namespace_values(namespace, order_by)
    [
      namespace.fetch("name").to_s,
      namespace.dig("resources_requested", order_by).to_i,
      namespace.dig("resources_used", order_by).to_i,
    ]
  end

  def namespaces
    @data.fetch("data")
  end
end
