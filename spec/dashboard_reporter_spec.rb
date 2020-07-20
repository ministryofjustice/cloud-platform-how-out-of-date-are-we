require "spec_helper"

describe DashboardReporter do

  let(:data) { {
    "updated_at" =>"2020-07-20 09:27:44",
    "data" => {
      "action_items" => {
        "documentation" => 1,
        "helm_whatup" => 1,
        "repositories" => 3,
        "terraform_modules" => 0
      },
      "action_required" => action_required
    }
  } }

  subject(:dr) { described_class.new }

  before do
    allow(dr).to receive(:fetch_data).and_return(data)
  end

  describe "action_required?" do
    context "when there are no open todo items" do
      let(:action_required) { false }

      it "reports nothing to do" do
        expect(dr.action_required?).to be false
      end
    end

    context "when there are open todo items" do
      let(:action_required) { true }

      it "reports something to do" do
        expect(dr.action_required?).to be true
      end
    end

    context "when data is incorrectly structured" do
      let(:data) { { "foo" => "bar" } }

      it "raises an error" do
        expect{
          dr.action_required?
        }.to raise_error(KeyError)
      end
    end
  end
end
