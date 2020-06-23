dev-server:
	API_KEY=soopersekrit ./app.rb -o 0.0.0.0

# These tests require the dev-server above to be running
test:
	rspec

fetch-live-json-datafiles:
	pod=$$(kubectl -n how-out-of-date-are-we get pods -o name); \
		for file in $$(kubectl -n how-out-of-date-are-we exec $${pod} ls data); do \
		  echo $${file}; \
			kubectl -n how-out-of-date-are-we exec $${pod} cat data/$${file} > data/$${file}; \
		done
