class LiveOneDomains
  def initialize(params)
    @data = JSON.parse(params.fetch(:json))
  end

  def total_ingress
  ingress.count
  end

  def updated_at
    DateTime.parse(@data.fetch("updated_at"))
  end

  def ingress
    @data.fetch("live_1_domains")
  end
end
