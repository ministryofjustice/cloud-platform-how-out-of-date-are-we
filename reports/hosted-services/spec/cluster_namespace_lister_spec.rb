RSpec.describe ClusterNamespaceLister do
  let(:name_a) { double(name: "aaa") }
  let(:name_b) { double(name: "bbb") }
  let(:name_ks) { double(name: "kube-system") }
  let(:name_knl) { double(name: "kube-node-lease") }

  let(:ns_a) { double(metadata: name_a) }
  let(:ns_b) { double(metadata: name_b) }
  let(:ns_ks) { double(metadata: name_ks) }
  let(:ns_knl) { double(metadata: name_knl) }

  let(:namespaces) { [ns_a, ns_b, ns_ks, ns_knl] }

  let(:params) {
    {
      config_file: "foo",
      context: "whatever",
    }
  }

  let(:kubeclient) { double(get_namespaces: namespaces) }
  let(:context) { double(api_endpoint: nil, ssl_options: nil, auth_options: nil) }
  let(:kubeconfig) { double(context: context) }

  subject(:lister) { described_class.new(params) }

  before do
    allow(Kubeclient::Config).to receive(:read).and_return(kubeconfig)
    allow(Kubeclient::Client).to receive(:new).and_return(kubeclient)
  end

  it "lists non-system namespace names" do
    expect(lister.namespace_names).to eq(["aaa", "bbb"])
  end

  it "does not include system namespaces" do
    expect(lister.namespaces).to_not include(ns_ks)
  end

  context "ingresses" do
    let(:system_ingress) {
      {
        "metadata" => {
          "namespace" => "kube-system",
        },
        "spec" => {
          "rules" => [
            "host" => "some.host.name",
          ],
        },
      }
    }

    let(:non_system_ingress) {
      {
        "metadata" => {
          "namespace" => "mynamespace",
        },
        "spec" => {
          "rules" => [
            "host" => "some.other.host",
          ],
        },
      }
    }

    let(:ingresses) {
      [
        system_ingress,
        non_system_ingress,
      ]
    }

    let(:json) {
      {"items" => ingresses}.to_json
    }

    let(:success) { double(Object, success?: true) }

    before do
      allow(Open3).to receive(:capture3).and_return([json, "", success])
    end

    it "lists ingresses in non-system namespaces" do
      expect(lister.ingresses).to eq([non_system_ingress])
    end
  end
end
