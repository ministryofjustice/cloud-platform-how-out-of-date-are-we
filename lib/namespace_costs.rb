class NamespaceCost
  def initialize(params)
    @file = params.fetch(:file)
  end

  def data
    @data ||= JSON.parse(File.read(@file))
  end

  def namespace
    File.basename(@file).sub(".json", "")
  end

  def updated_at
    File.stat(@file).mtime
  end

  def resources
    @resources ||= begin
                     data.map do |resource|
                       {
                         name: resource["name"],
                         monthly_cost: resource["monthlyCost"].to_f
                       }
                     end
                   end
  end

  # file contains the output of `infracost --tfdir .` for each namespace
  def total
    @total ||= data.map { |item| item["monthlyCost"].to_f }.sum
  end
end

class NamespaceCosts
  attr_reader :dir

  def initialize(params)
    @dir = params.fetch(:dir)
  end

  def updated_at
    @updated_at ||= list.map(&:updated_at).min
  end

  def list
    @list ||= Dir["#{dir}/*.json"]
      .map { |file| NamespaceCost.new(file: file) }
      .sort {|a,b| a.total <=> b.total}
      .reverse
  end
end
