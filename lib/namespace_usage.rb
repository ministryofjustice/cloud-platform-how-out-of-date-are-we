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
    namespaces.sum { |n| n.dig("Used", "Pods") }
  end

  def updated_at
    DateTime.parse(@data.fetch("updated_at"))
  end

  def total_requested(property)
    namespaces.map { |ns| ns.dig("Requested", property) }.map(&:to_i).sum
  end

  def namespace(name)
    namespaces.find { |n| n["Name"] == name }
  end

  private

  def namespace_pods_values(namespace)
    [
      namespace.fetch("Name").to_s,
      namespace.dig("Hardlimits", "Pods").to_i,
      namespace.dig("Used", "Pods").to_i,
    ]
  end

  def namespace_values(namespace, order_by)
    [
      namespace.fetch("Name").to_s,
      namespace.dig("Requested", order_by).to_i,
      namespace.dig("Used", order_by).to_i,
    ]
  end

  def namespaces
    @data.fetch("data")
  end
end
