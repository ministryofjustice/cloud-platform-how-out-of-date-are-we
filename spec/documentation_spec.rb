require "spec_helper"

describe Documentation do
  let(:json) { data.to_json }
  let(:logger) { double(Sinatra::CommonLogger) }

  let(:params) { {
    file: "foo.json",
    key: key,
    logger: logger,
  } }

  let(:key) { "pages" }
  let(:site) { "runbooks" }
  let(:title) { "joiners-guide" }
  let(:url) { "https://#{site}.cloud-platform.service.justice.gov.uk/#{title}.html" }
  let(:pages) { [ url ] }
  let(:data) {
    {
      pages: pages,
      updated_at: "2020-07-14 12:34:56",
    }
  }

  subject(:documentation) { described_class.new(params) }

  before do
    allow(FileTest).to receive(:exists?).and_return(true)
    allow(File).to receive(:read).and_return(json)
    allow(logger).to receive(:info)
  end

  it "converts url list to list of tuples" do
    expected = [ { "site" => site, "title" => title, "url" => url } ]
    expect(documentation.list).to eq(expected)
  end
end
