class HelmWhatup < ItemList
  DANGER = "danger"
  WARNING = "warning"
  SUCCESS = "success"

  def out_of_date_apps
    helm_charts.filter { |chart| version_lag_traffic_light(chart) == DANGER }
  end

  def todo_count
    out_of_date_apps.length
  end

  private

  def helm_charts
    list.map { |cluster| cluster.fetch("apps") }.flatten
  end

  def read_data
    data = super

    data&.fetch("clusters")&.each do |cluster|
      cluster.fetch("apps").map { |app| app["traffic_light"] = version_lag_traffic_light(app) }
    end

    data
  end

  # Return success/warning/danger, depending on
  # how far behind latest the installed version
  # is.
  def version_lag_traffic_light(app)
    installed = app.fetch("installed_version").split(".")
    latest = app.fetch("latest_version").split(".")

    major_diff = latest[0].to_i - installed[0].to_i
    minor_diff = latest[1].to_i - installed[1].to_i

    if major_diff > 1
      DANGER
    elsif minor_diff > 4
      WARNING
    else
      SUCCESS
    end
  end
end
