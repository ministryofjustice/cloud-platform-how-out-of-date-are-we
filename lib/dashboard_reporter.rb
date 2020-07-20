require "open-uri"
require "json"

class DashboardReporter
  attr_reader :dashboard_url

  def initialize(url)
    @dashboard_url = url
  end

  def slack_formatted_report
    return "" unless action_required?

    items = data.fetch("data").fetch("action_items")
    item_keys = items.keys.sort
    max_key_length = item_keys.map(&:length).max

    action_items = item_keys.map do |key|
      value = items[key]
      label = "#{key}:".ljust(max_key_length + 1)
      [label, value].join(" ")
    end

    %(
Action items:
```
#{action_items.join("\n")}
```
    ).strip
  end

  private

  def action_required?
    data.fetch("data").fetch("action_required")
  end

  def data
    @data ||= begin
      json = URI.open(dashboard_url, "Accept" => "application/json") { |f| f.read }
      JSON.parse(json)
    end
  end
end
