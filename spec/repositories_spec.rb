require "spec_helper"

describe GithubRepositories do
  let(:json) { data.to_json }
  let(:logger) { double(Sinatra::CommonLogger) }

  let(:params) {
    {
      file: "foo.json",
      key: key,
      logger: logger,
    }
  }

  let(:key) { "repositories" }
  let(:data) {
    {
      repositories: repositories,
      updated_at: "2020-07-14 12:34:56",
    }
  }

  let(:pass_repos) {
    [
      {"status" => "PASS", "foo" => "bar"},
      {"status" => "PASS", "foo" => "baz"},
    ]
  }

  let(:fail_repos) { [{"status" => "NOTPASS", "foo" => "xxx"}] }

  let(:repositories) { pass_repos + fail_repos }

  subject(:repos) { described_class.new(params) }

  before do
    allow(FileTest).to receive(:exists?).and_return(true)
    allow(File).to receive(:read).and_return(json)
    allow(logger).to receive(:info)
  end

  it "excludes passing repos from list" do
    expect(repos.list).to eq(fail_repos)
  end
end
