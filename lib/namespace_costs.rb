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
    @updated_at ||= list.map {|i| i.fetch(:updated_at)}.min
  end

  def list
    @list ||=
      begin
        Dir["#{dir}/*.json"]
          .sort
          .map do |file|
            nc = NamespaceCost.new(file: file)
            {
              namespace: nc.namespace,
              total: nc.total,
              updated_at: nc.updated_at,
            }
          end
      end
  end
end
