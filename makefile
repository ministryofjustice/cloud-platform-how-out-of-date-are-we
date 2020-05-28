dev-server:
	API_KEY=soopersekrit ./app.rb -o 0.0.0.0

# These tests require the dev-server above to be running
test:
	rspec
