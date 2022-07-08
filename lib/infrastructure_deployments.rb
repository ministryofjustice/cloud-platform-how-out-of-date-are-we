class InfraDeployments
  def initialize(params)
    @data = JSON.parse(params.fetch(:json))
  end

  def updated_at
    DateTime.parse(@data.fetch("updated_at"))
  end

  def deployments
    @data.fetch("deployments")
  end
end
