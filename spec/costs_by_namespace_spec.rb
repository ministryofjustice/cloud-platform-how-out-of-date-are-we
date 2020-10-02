require "spec_helper"

describe CostsByNamespace do
  let(:updated_at) { "2020-10-02 16:24:25" }

  let(:data) { {
    "namespace" => {
      "SHARED" => {
        "breakdown" => {},
        "total" => 33.33
      },
      "aaa" => {
        "breakdown" => {},
        "total" => 11.11
      },
      "bbb" => {
        "breakdown" => {},
        "total" => 22.22
      },
    },
    "updated_at" => updated_at
  } }

  let(:json) { data.to_json }

  subject(:cbn) { described_class.new(json: json) }

  specify { expect(cbn.updated_at).to eq(updated_at) }
  specify { expect(cbn.list).to be_an(Array) }

  it "calculates total" do
    total = 11.11 + 22.22 + 33.33
    expect(cbn.total).to eq(total)
  end

  it "orders by descending total" do
    names = cbn.list.map { |i| i["name"] }
    expect(names).to eq(["SHARED", "bbb", "aaa"])
  end

  it "adds name to namespace hash" do
    name = cbn.list.first["name"]
    expect(name).to eq("SHARED")
  end
end
