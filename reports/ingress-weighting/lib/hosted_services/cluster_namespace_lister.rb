class ClusterNamespaceLister
  attr_reader :config_file, :context

  K8S_DEFAULT_NAMESPACES = %w[
    cert-manager
    default
    ingress-controllers
    kiam
    kube-node-lease
    kube-public
    kube-system
    kuberos
    opa
    velero
  ]

  def initialize(args)
    @config_file = args.fetch(:config_file)
    @context = args.fetch(:context)
  end

  def kubeclient
    kubeconfig = Kubeclient::Config.read(config_file)
    ctx = kubeconfig.context(context)
    Kubeclient::Client.new(
      ctx.api_endpoint,
      "v1",
      ssl_options: ctx.ssl_options,
      auth_options: ctx.auth_options
    )
  end

  def namespace_names
    namespaces.map { |n| n.metadata.name }
  end

  def namespaces
    kubeclient
      .get_namespaces
      .reject { |n| K8S_DEFAULT_NAMESPACES.include?(n.metadata.name) }
  end

  def ingresses
    kubectl_get_ingresses
      .reject { |i| K8S_DEFAULT_NAMESPACES.include?(i.dig("metadata", "namespace")) }
  end

  private

  def kubectl_get_ingresses
    cmd = [
      "kubectl config use-context #{context} > /dev/null", # So we don't get "Switched to context..." in the output
      "kubectl get ingresses --all-namespaces -o json",
    ].join("; ")

    stdout, stderr, status = Open3.capture3(cmd)

    unless status.success?
      raise stderr
    end

    JSON.parse(stdout).fetch("items")
  end
end
