helpers do
  # Return success/warning/danger, depending on
  # how far behind latest the installed version
  # is.
  def version_lag_traffic_light(app)
    installed = app.fetch("installedVersion").split(".")
    latest = app.fetch("latestVersion").split(".")

    major_diff = latest[0].to_i - installed[0].to_i
    minor_diff = latest[1].to_i - installed[1].to_i

    if major_diff > 1
      "danger"
    elsif minor_diff > 4
      "warning"
    else
      "success"
    end
  end

  def correct_api_key?(request)
    expected_key = ENV.fetch("API_KEY")
    provided_key = request.env.fetch("HTTP_X_API_KEY", "dontsetthisvalueastheapikey")

    expected_key == provided_key
  end
end
