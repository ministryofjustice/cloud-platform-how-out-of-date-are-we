# Fetch AWS costs for a single day, broken down by namespace tag, and multiply
# all amounts by 30 to get monthly costs
class AwsCostsByNamespace
  attr_reader :date, :access_key_id, :secret_access_key

  TAG = "namespace"
  SHARED_COSTS = "SHARED_COSTS"
  DAYS_PER_MONTH = 30 # An average, to convert daily amounts to monthly

  # Annual cost of the Cloud Platform team is £1,260,000
  # This is the monthly cost in USD
  MONTHLY_TEAM_COST = 136_000

  REGION = "us-east-1" # AWS CostExplorer only works with this region value

  def initialize(params)
    @date = params.fetch(:date) # should be a Date object
    @access_key_id = params.fetch(:access_key_id, ENV.fetch("AWS_ACCESS_KEY_ID"))
    @secret_access_key = params.fetch(:secret_access_key, ENV.fetch("AWS_SECRET_ACCESS_KEY"))
  end

  def report
    tuples = aws_data.map { |cost| hash_from_cost(cost) }
    costs = costs_by_namespace(tuples)
    add_shared_aws_costs(costs)
    add_shared_team_cost(costs)
    add_totals(costs)

    {
      TAG => costs,
      :updated_at => Time.now
    }
  end

  private

  # Add a share of the monthly team cost to each namespace
  def add_shared_team_cost(costs)
    cost_per_namespace = MONTHLY_TEAM_COST.to_f / costs.keys.size
    costs.values.each { |hash| hash["Shared CP Team Costs"] = cost_per_namespace }
  end

  # Extract total shared AWS cost and divide it evenly over namespaces
  def add_shared_aws_costs(costs)
    shared_costs = costs.delete(SHARED_COSTS).values.sum
    cost_per_namespace = shared_costs / costs.keys.size
    costs.values.each { |hash| hash["Shared AWS Costs"] = cost_per_namespace }
  end

  # Take output from AWS CostExplorer and convert to a hash
  #   { [namespace name] => {
  #     "breakdown" => {
  #       [resource type] => cost,
  #       ...
  #     },
  #     "total" => [monthly amount]
  #   }
  def costs_by_namespace(tuples)
    tuples.each_with_object({}) do |h, acc|
      tag = h[:tag]
      resource = h[:resource]
      costs = acc[tag] || {}
      costs[resource] = costs[resource].to_f + h[:amount]
      acc[tag] = costs
    end
  end

  # { foo: { a: 1, b: 2 } } -> { foo: { breakdown: { a: 1, b: 2 }, total: 3 } }
  def add_totals(costs)
    costs.each do |namespace, resource_costs|
      costs[namespace] = {
        breakdown: resource_costs,
        total: resource_costs.values.sum
      }
    end
  end

  def hash_from_cost(cost)
    resource_type, tag_string = cost.keys
    tag_value = tag_string.split("$")[1].to_s
    tag_value = SHARED_COSTS if tag_value == ""

    {
      resource: resource_type,
      tag: tag_value,
      amount: cost.metrics.fetch("BlendedCost").amount.to_f * DAYS_PER_MONTH
    }
  end

  def aws_data
    ce = Aws::CostExplorer::Client.new(
      region: REGION,
      access_key_id: access_key_id,
      secret_access_key: secret_access_key
    )

    end_date = date.strftime("%Y-%m-%d")
    start_date = date.prev_day.strftime("%Y-%m-%d")

    data = ce.get_cost_and_usage(
      granularity: "DAILY",
      metrics: ["BlendedCost"],
      time_period: {
        start: start_date,
        end: end_date
      },
      group_by: [
        {
          type: "DIMENSION",
          key: "SERVICE"
        },
        {
          type: "TAG",
          key: TAG
        }
      ]
    )

    raise "More than one page in response. Please add code to iterate over all response pages." if data.next_page?
    raise "More than one entry in results_by_time - I don't know how to handle that." if data.results_by_time.size > 1

    data.results_by_time.first.groups
  end
end
