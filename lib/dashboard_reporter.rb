require "open-uri"
require "json"

class DashboardReporter
  attr_reader :dashboard_url

  def initialize(url)
    @dashboard_url = url
  end

  def action_required?
    data.fetch("data").fetch("action_required")
  end

  private

  def data
    @data ||= begin
      json = URI.open(url, "Accept" => "application/json") { |f| f.read }
      JSON.parse(json)
    end
  end
end
