class NamespaceCosts
  attr_reader :dir

  def initialize(params)
    @dir = params.fetch(:dir)
  end

  def updated_at
    list # sets @updated_at as a side-effect
    @updated_at
  end

  def set_updated_at(file)
    @updated_at ||= Time.now
    @updated_at = [
      @updated_at,
      File.stat("data/namespace/costs/cccd-dev.json").mtime
    ].min
  end

  def list
    @list ||= Dir["#{dir}/*.json"]
      .sort
      .map { |file| process(file) }
  end

  private

  # file contains the output of `infracost --tfdir .` for each namespace
  def process(file)
    items = JSON.parse(File.read(file))
    set_updated_at(file)
    total = items.map { |item| item["monthlyCost"].to_f }.sum
    namespace = File.basename(file).sub(".json", "")
    {
      name: namespace,
      total: total,
    }
  end
end
