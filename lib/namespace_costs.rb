class NamespaceCosts
  attr_reader :dir

  def initialize(params)
    @dir = params.fetch(:dir)
  end

  def updated_at
    Time.now # TODO: oldest json file modification date
  end

  def list
    []
  end
end
