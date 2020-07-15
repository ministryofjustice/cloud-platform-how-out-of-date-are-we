helpers do
  def correct_api_key?(request)
    expected_key = ENV.fetch("API_KEY")
    provided_key = request.env.fetch("HTTP_X_API_KEY", "dontsetthisvalueastheapikey")

    expected_key == provided_key
  end
end
