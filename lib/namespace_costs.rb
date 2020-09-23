class NamespaceCost
  def initialize(params)
    @store = params.fetch(:store)
    @file = params.fetch(:file)
    @data = params[:data] # used when recreating from cached JSON data
    @updated_at = params[:updated_at] # used when recreating from cached JSON data
  end

  def data
    @data ||= JSON.parse(@store.retrieve_file(@file))
  end

  def namespace
    File.basename(@file).sub(".json", "")
  end

  def updated_at
    @updated_at ||= @store.stored_at(@file)
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

  def to_json(_)
    {
      file: @file,
      data: data,
      updated_at: updated_at,
    }.to_json
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
    cost_files = store.list_files.filter { |f| f =~ %r[^#{dir}/.*.json$] }

    list = store.retrieve_files(cost_files).map do |file, hash|
      NamespaceCost.new(
        file: file,
        data: JSON.parse(hash["content"]),
        updated_at: hash["stored_at"],
        store: nil,
      )
    end

    list
      .sort {|a,b| a.total <=> b.total}
      .reverse
  end
end
