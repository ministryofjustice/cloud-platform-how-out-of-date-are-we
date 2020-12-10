require "spec_helper"

describe "orphaned statefiles" do
  let(:json) { data.to_json }
  let(:logger) { double(Sinatra::CommonLogger) }

  let(:params) {
    {
      file: "foo.json",
      key: "data",
      logger: logger,
    }
  }

  let(:data) {
    {
      "data" => [
        "file1",
        "file2",
      ]
    }
  }

  subject(:orphaned_statefiles) { ItemList.new(params) }

  before do
    allow(FileTest).to receive(:exists?).and_return(true)
    allow(File).to receive(:read).and_return(json)
    allow(logger).to receive(:info)
  end

  it "counts all items" do
    expect(orphaned_statefiles.todo_count).to eq(2)
  end
end
