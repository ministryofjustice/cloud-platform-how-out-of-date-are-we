class DashboardReporter
  def action_required?
    fetch_data.fetch("data").fetch("action_required")
  end

  def fetch_data
    @data ||= {}
  end
end
