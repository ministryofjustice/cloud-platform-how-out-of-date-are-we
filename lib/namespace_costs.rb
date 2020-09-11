class NamespaceCosts
  attr_reader :dir

  def initialize(params)
    @dir = params.fetch(:dir)
  end

  def updated_at
    Time.now # TODO: oldest json file modification date
  end

  def list
    Dir["#{dir}/*.json"].sort.map { |file| process(file) }
  end

  private

  # file contains the output of `infracost --tfdir .` for each namespace
  def process(file)
    items = JSON.parse(File.read(file))
    total = items.map { |item| item["monthlyCost"].to_f }.sum
    namespace = File.basename(file).sub(".json", "")
    {
      name: namespace,
      total: total,
    }
  end
end
