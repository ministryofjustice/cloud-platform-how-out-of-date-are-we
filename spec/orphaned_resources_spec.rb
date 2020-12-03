require "spec_helper"

describe OrphanedResources do
  let(:json) { data.to_json }
  let(:logger) { double(Sinatra::CommonLogger) }

  let(:params) {
    {
      file: "foo.json",
      key: key,
      logger: logger,
    }
  }

  let(:key) { "orphaned_aws_resources" }

  let(:data) {
    {
      "orphaned_aws_resources" => {
        "hosted_zones" => [
          {"id" => "z1", "cluster" => ""},
          {"id" => "z2", "cluster" => ""},
          {"id" => "z3", "cluster" => ""},
        ],
        "vpcs" => [
          {"id" => "v1", "cluster" => "test-1"},
          {"id" => "v2", "cluster" => "test-2"},
        ],
      },
    }
  }

  subject(:orphaned_resources) { described_class.new(params) }

  before do
    allow(FileTest).to receive(:exists?).and_return(true)
    allow(File).to receive(:read).and_return(json)
    allow(logger).to receive(:info)
  end

  it "counts all items" do
    expect(orphaned_resources.todo_count).to eq(5)
  end
end
