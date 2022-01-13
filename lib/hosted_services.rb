class HostedServices
  def initialize(params)
    @data = JSON.parse(params.fetch(:json))
  end

  def unique_apps
    namespaces.map { |h| h["Application"] }.uniq.sort.count
  end

  def total_ns
    namespaces.count
  end

  def updated_at
    DateTime.parse(@data.fetch("updated_at"))
  end

  def namespaces
    @data.fetch("namespace_details")
  end
end
