require "spec_helper"

describe HelmWhatup do
  let(:json) { data.to_json }
  let(:logger) { double(Sinatra::CommonLogger) }

  let(:params) { {
    file: "foo.json",
    key: key,
    logger: logger,
  } }

  let(:key) { "clusters" }

  let(:red) { {
    "name" => "red",
    "installed_version" => "1.2.3",
    "latest_version" => "5.0.0",
  } }

  let(:yellow) { {
    "name" => "yellow",
    "installed_version" => "1.2.3",
    "latest_version" => "1.3.0",
  } }

  let(:green) { {
    "name" => "green",
    "installed_version" => "1.2.3",
    "latest_version" => "1.2.4",
  } }

  let(:data) {
    {
      clusters: [
        {
          name: "manager",
          apps: [ red, yellow ],
        },
        {
          name: "live-1",
          apps: [ green ],
        },
      ],
      updated_at: "2020-07-14 12:34:56",
    }
  }

  subject(:helm_whatup) { described_class.new(params) }

  before do
    allow(FileTest).to receive(:exists?).and_return(true)
    allow(File).to receive(:read).and_return(json)
    allow(logger).to receive(:info)
  end

  it "counts red apps as todos" do
    expect(helm_whatup.todo_count).to eq(1)
  end

  it "returns red apps" do
    expected = [ red.merge("traffic_light" => "danger") ]
    expect(helm_whatup.out_of_date_apps).to eq(expected)
  end
end
