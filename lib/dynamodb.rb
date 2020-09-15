class Dynamodb
  attr_reader :db, :table

  def initialize(params)
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

  def store_file(file)
    puts file
    json = File.read(file)
    db.put_item(table_name: table, item: { filename: file, content: json})
  end

  def retrieve_file(file)
    db.get_item(
      key: { filename: file },
      table_name: table,
    ).item["content"]
  end
end
