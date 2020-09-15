class Dynamodb
  attr_reader :db, :table

  def initialize(params = {})
    @db = params.fetch(:db, Aws::DynamoDB::Client.new(
      region: ENV.fetch("DYNAMODB_REGION"),
      access_key_id: ENV.fetch("DYNAMODB_ACCESS_KEY_ID"),
      secret_access_key: ENV.fetch("DYNAMODB_SECRET_ACCESS_KEY"),
    ))
    @table = params.fetch(:table, ENV.fetch("DYNAMODB_TABLE_NAME"))
  end

  def list_files
    db.scan(
      table_name: table,
      expression_attribute_names: { "#F" => "filename" },
      projection_expression: "#F",
    ).items.map { |i| i["filename"] }.sort
  end

  def store_file(file, content)
    db.put_item(table_name: table, item: { filename: file, content: content})
  end

  def retrieve_file(file)
    item = get_item(file)
    item.nil? ? nil : item["content"]
  end

  def stored_at(file)
    exists?(file) ? Time.parse(get_item(file)["stored_at"]) : nil
  end

  def exists?(file)
    !retrieve_file(file).nil?
  end

  private

  def get_item(key)
    db.get_item(
      key: { filename: key },
      table_name: table,
    ).item
  end
end
