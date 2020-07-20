require "open-uri"
require "json"

class DashboardReporter
  attr_reader :dashboard_url

  def initialize(url)
    @dashboard_url = url
  end

  def report
    return "" unless action_required?

    %(
How out of date are we - action required:
```
documentation:     1
helm_whatup:       1
repositories:      3
terraform_modules: 0
```
    ).strip
  end

  private

  def action_required?
    data.fetch("data").fetch("action_required")
  end

  def data
    @data ||= begin
      json = URI.open(url, "Accept" => "application/json") { |f| f.read }
      JSON.parse(json)
    end
  end
end
