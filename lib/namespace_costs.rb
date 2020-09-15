class NamespaceCost
  def initialize(params)
    @store = params.fetch(:store)
    @file = params.fetch(:file)
  end

  def data
    @data ||= JSON.parse(@store.retrieve_file(@file))
  end

  def namespace
    File.basename(@file).sub(".json", "")
  end

  def updated_at
    @store.stored_at(@file)
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
  attr_reader :dir, :store

  def initialize(params)
    @store = params.fetch(:store)
    @dir = params.fetch(:dir)
  end

  def updated_at
    @updated_at ||= list.map(&:updated_at).min
  end

  def list
    @list ||= costs_list
  end

  def total
    list.sum(&:total)
  end

  private

  def costs_list
    store.list_files
      .filter { |f| f =~ %r[^#{dir}/.*.json$] }
      .map { |file| NamespaceCost.new(file: file, store: store) }
      .sort {|a,b| a.total <=> b.total}
      .reverse
  end
end
