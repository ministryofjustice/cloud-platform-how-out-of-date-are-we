############################################################

url = ENV.fetch("DASHBOARD_URL")
output_file = ENV.fetch("OUTPUT_FILE")
timestamp = Time.now.strftime("%Y-%m-%d %H:%M:%S")

report = [DashboardReporter.new(url).slack_formatted_report, timestamp].join("\n")

if report == ""
  puts "#{timestamp} No action items reported."
else
  puts report # <-- so we see it in the concourse log
  File.open(output_file, "w") { |f| f.puts report }
  exit 1
end
