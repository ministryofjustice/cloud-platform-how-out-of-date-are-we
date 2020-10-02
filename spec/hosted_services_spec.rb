require "spec_helper"

describe "hosted_services" do
  let(:json) { data.to_json }
  let(:logger) { double(Sinatra::CommonLogger) }

  let(:params) { {
    file: "foo.json",
    key: key,
    logger: logger,
  } }

  let(:key) { "namespace_details" }
  let(:data) {
    {
      "namespace_details" => namespace_details,
      "updated_at" => "2020-07-14 12:34:56",
    }
  }

  let(:namespace_details) { [
        {
          "namespace" => "mynamespace1", 
          "application" => "MyApplication1",
          "business_unit" => "BusinessUnit1",
          "team_name" => "team1",
          "team_slack_channel" => "slack-channel-team1",
          "github_url" => "https://www.example1.com",
          "deployment_type" => "prod1",
          "domain_names" => ["www.domain1.com"],
        },
        {
          "namespace" => "mynamespace2", 
          "application" => "MyApplication2",
          "business_unit" => "BusinessUnit2",
          "team_name" => "team2",
          "team_slack_channel" => "slack-channel-team2",
          "github_url" => "https://www.example2.com",
          "deployment_type" => "prod2",
          "domain_names" => ["www.domain2.com"],
        }
      ]
  }

  subject(:hosted_services) { ItemList.new(params) }

  before do
    allow(FileTest).to receive(:exists?).and_return(true)
    allow(File).to receive(:read).and_return(json)
    allow(logger).to receive(:info)
  end

  it "counts all items" do
    expect(hosted_services.todo_count).to eq(2)
  end
end
