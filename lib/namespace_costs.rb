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

  # Generating the list requires a lot of file lookups, which is slow if we're using
  # dynamodb, so we cache the generated list in this file.
  CACHE_FILENAME = "data/namespace_costs.json"
  CACHE_SECONDS = 3600

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
    if cache_expired?
      data = {
        "list" => build_costs_list,
        "updated_at" => Time.now,
      }
      cache(data)
    else
      data = cached_costs_list
    end
    data["list"]
  end

  def cache_expired?
    if store.exists?(CACHE_FILENAME)
      (store.stored_at(CACHE_FILENAME) + CACHE_SECONDS) < Time.now
    else
      true
    end
  end

  def cached_costs_list
    data = JSON.parse(store.retrieve_file CACHE_FILENAME)

    list = data["list"].map { |d|
      NamespaceCost.new(
        file: d["file"],
        data: d["data"],
        updated_at: d["updated_at"],
        store: nil,
      )
    }

    {
      "list" => list,
      "updated_at" => data["updated_at"],
    }
  end

  def cache(data)
    store.store_file(CACHE_FILENAME, data.to_json)
  end

  def build_costs_list
    store.list_files
      .filter { |f| f =~ %r[^#{dir}/.*.json$] }
      .map { |file| NamespaceCost.new(file: file, store: store) }
      .sort {|a,b| a.total <=> b.total}
      .reverse
  end
end
