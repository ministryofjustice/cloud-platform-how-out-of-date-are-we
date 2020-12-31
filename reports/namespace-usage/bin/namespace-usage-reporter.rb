#!/usr/bin/env ruby

# This script outputs a report to assist in comparing the resources a namespace
# is requesting (and allowed to request), against what it actually uses.

require "json"

def main
  r = NamespaceReporter.new
  puts({data: r.reports, updated_at: Time.now}.to_json)
end

class NamespaceReporter
  def namespaces
    `kubectl get ns -o jsonpath='{.items[*].metadata.name}'`.chomp
      .split(" ")
  end

  def reports
    namespaces.map { |namespace| report(namespace) }
  end

  private

  def report(namespace)
    ns_quota = quota(namespace)
    ns_limits = limits(namespace)
    pod_data = pods(namespace)

    {
      name: namespace,
      resources_used: top_pods(namespace),
      default_request: default_request(ns_limits),
      default_limit: default_limit(ns_limits),
      max_requests: ns_quota.fetch(:hard_request_limit),
      hard_limit: ns_quota.fetch(:hard_limit),
      hard_limit_used: ns_quota.fetch(:hard_limit_used),
      resources_requested: resources_requested(pod_data),
      container_count: container_count(pod_data)
    }
  end

  def quota(namespace)
    quota = all_quotas.find { |q| q.dig("metadata", "namespace") == namespace }

    if quota.nil?
      {
        hard_request_limit: {cpu: nil, memory: nil},
        hard_limit: {cpu: nil, memory: nil, pods: nil},
        requested: {cpu: nil, memory: nil},
        hard_limit_used: {cpu: nil, memory: nil}
      }
    else
      data = quota.dig("status")

      hard_request_limit = {
        cpu: Num.cpu_value(data.dig("hard", "requests.cpu")),
        memory: Num.memory_value(data.dig("hard", "requests.memory"))
      }

      hard_limit = {
        cpu: Num.cpu_value(data.dig("hard", "limits.cpu")),
        memory: Num.memory_value(data.dig("hard", "limits.memory")),
        pods: Num.integer_value(data.dig("hard", "pods"))
      }

      hard_limit_used = {
        cpu: Num.cpu_value(data.dig("used", "limits.cpu")),
        memory: Num.memory_value(data.dig("used", "limits.memory"))
      }

      requested = {
        cpu: Num.cpu_value(data.dig("used", "requests.cpu")),
        memory: Num.memory_value(data.dig("used", "requests.memory"))
      }

      {
        hard_request_limit: hard_request_limit,
        hard_limit: hard_limit,
        hard_limit_used: hard_limit_used,
        requested: requested
      }
    end
  end

  def all_quotas
    @all_quotas ||= kubectl_get("quota")
  end

  def limits(namespace)
    all_limits.find { |l| l.dig("metadata", "namespace") == namespace }
  end

  def all_limits
    @all_limits ||= kubectl_get("limits")
  end

  def pods(namespace)
    all_pods.filter { |p| p.dig("metadata", "namespace") == namespace }
  end

  def all_pods
    @all_pods ||= kubectl_get("pods")
  end

  def top_pods(namespace)
    @top_pods ||= TopPods.new
    @top_pods.for_namespace(namespace)
  end

  def default_request(limits)
    from_limits(limits, "defaultRequest")
  end

  def default_limit(limits)
    from_limits(limits, "default")
  end

  def resources_requested(data)
    cpu = data.inject(0) { |sum, item|
      requested = item.dig("spec", "containers").inject(0) { |total, container|
        total += Num.cpu_value(container.dig("resources", "requests", "cpu")).to_i
      }
      sum += requested
    }

    memory = data.inject(0) { |sum, item|
      requested = item.dig("spec", "containers").inject(0) { |total, container|
        total += Num.memory_value(container.dig("resources", "requests", "memory")).to_i
      }
      sum += requested
    }

    {cpu: cpu, memory: memory, pods: data.count}
  end

  def from_limits(limits, value_type)
    data = limits.dig("spec", "limits")[0]
      .dig(value_type)

    {
      cpu: cpu_value(data.fetch("cpu", nil)),
      memory: memory_value(data.fetch("memory", nil))
    }
  rescue
    {
      cpu: nil,
      memory: nil
    }
  end

  def container_count(data)
    data.collect { |i| i.dig("spec", "containers") }.flatten.compact.count
  end

  def kubectl_get(obj)
    JSON.parse(`kubectl get #{obj} --all-namespaces -o json`)
      .fetch("items")
  end
end

class TopPods
  def for_namespace(namespace)
    data[namespace] || {cpu: 0, memory: 0, pods: 0}
  end

  private

  def data
    @data ||= parse(raw_data)
  end

  def parse(txt)
    lines = txt.split("\n")
    lines.shift

    tuples = lines.map { |l|
      namespace, pod, cpu, memory = l.split(" ")
      {
        namespace: namespace,
        pod: pod,
        cpu: cpu,
        memory: memory
      }
    }

    tuples.each_with_object({}) do |t, hash|
      namespace = t[:namespace]
      hash[namespace] ||= {cpu: 0, memory: 0, pods: 0}
      hash[namespace][:cpu] += Num.cpu_value(t[:cpu])
      hash[namespace][:memory] += Num.memory_value(t[:memory])
      hash[namespace][:pods] += 1
    end
  end

  def raw_data
    @raw_data ||= `kubectl top pod --all-namespaces`
  end
end

class Num
  def self.cpu_value(str)
    return nil if str.nil?

    case str
    when /^(\d+)$/
      $1.to_i * 1000
    when /^(\d+)m$/
      $1.to_i
    else
      raise %(CPU value: "#{str}" was not in expected format)
    end
  end

  def self.integer_value(str)
    str.nil? ? nil : str.to_i
  end

  def self.memory_value(str)
    return nil if str.nil?

    case str
    when /^(\d+)$/
      $1.to_i / 1_000
    when /^(\d+)k$/, /^(\d+)Ki$/
      $1.to_i / 1024
    when /^(\d+)m$/ # e.g. 6.4Gi in yaml => 6871947673600m in the JSON kubectl output
      $1.to_i / 1_000_000_000
    when /^(\d+)Gi/
      $1.to_i * 1000
    when /^(\d+)Mi/
      $1.to_i
    else
      raise %(Memory value: "#{str}" was not in expected format)
    end
  end
end

main
