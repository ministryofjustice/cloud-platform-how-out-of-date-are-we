require "spec_helper"

# These specs test a local development server, as started via
#
#     make dev-server
#

HELM_RELEASE_DATA_FILE = "data/helm_whatup.json"

describe "local dev server" do
  let(:base_url) { "http://localhost:4567" }
  let(:api_key) { "soopersekrit" } # specified in makefile

  let(:helm_whatup_url) { [base_url, "helm_whatup"].join("/") }
  let(:terraform_modules_url) { [base_url, "terraform_modules"].join("/") }
  let(:documentation_url) { [base_url, "documentation"].join("/") }

  let(:urls) { [
    helm_whatup_url,
    terraform_modules_url,
    documentation_url,
  ] }

  it "redirects / to /helm_whatup" do
    response = fetch_url(base_url)
    expect(response.code).to eq("302")
    expect(response["location"]).to eq(helm_whatup_url)
  end

  it "serves pages" do
    urls.each do |url|
      response = fetch_url(url)
      expect(response.code).to eq("200")
    end
  end

  context "with malformed json data" do
    before do
      File.write(HELM_RELEASE_DATA_FILE, " ")
    end

    it "does not crash" do
      response = fetch_url(helm_whatup_url)
      expect(response.code).to eq("200")
    end
  end

  context "updating" do
    context "with no API key" do
      it "rejects" do
        urls.each do |url|
          response = post_to_url(url, "[]")
          expect(response.code).to eq("403")
        end
      end
    end

    context "with correct API key" do
      it "accepts helm_whatup json" do
        json = {
          clusters: [
            name: "live-1",
            apps: []
          ],
          updated_at: Time.now
        }.to_json
        response = post_to_url(helm_whatup_url, json, api_key)
        expect(response.code).to eq("200")
      end

      it "accepts terraform_modules json" do
        json = {
          out_of_date_modules: [],
          updated_at: "2020-04-20",
        }.to_json

        response = post_to_url(terraform_modules_url, json, api_key)
        expect(response.code).to eq("200")
      end

      it "accepts documentation json" do
        json = {
          pages: [],
          updated_at: "2020-04-20",
        }.to_json

        response = post_to_url(documentation_url, json, api_key)
        expect(response.code).to eq("200")
      end
    end

    context "with incorrect API key" do
      let(:api_key) { "this is the wrong API key" }

      it "rejects" do
        urls.each do |url|
          response = post_to_url(url, "[]", api_key)
          expect(response.code).to eq("403")
        end
      end
    end
  end
end
