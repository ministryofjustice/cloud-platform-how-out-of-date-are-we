class ItemList
  attr_reader :logger

  def initialize(params)
    @json_file = params.fetch(:file)
    @key = params.fetch(:key) # top-level key containing list of items
    @logger = params.fetch(:logger)
  end

  def list
    @list ||= data.nil? ? [] : data.fetch(@key)
  end

  def updated_at
    @updated_at ||= data.nil? ? "" : string_to_formatted_time(data.fetch("updated_at"))
  end

  def todo_count
    list.length
  end

  private

  def data
    @data ||= read_data
  end

  def read_data
    data = nil

    unless FileTest.exists?(@json_file)
      logger.info "No such file #{@json_file}"
    else
      begin
        data = JSON.parse(File.read @json_file)
      rescue JSON::ParserError
        logger.info "Malformed JSON file: #{@json_file}"
      end
    end

    data
  end

  def string_to_formatted_time(str)
    DateTime.parse(str).strftime("%Y-%m-%d %H:%M:%S")
  end
end
