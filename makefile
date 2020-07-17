IMAGE := ministryofjustice/cloud-platform-how-out-of-date-are-we:2.11

dev-server:
	API_KEY=soopersekrit ./app.rb -o 0.0.0.0

docker-dev-server:
	docker run --rm \
		-v $$(pwd)/data:/app/data \
		-e API_KEY=soopersekrit \
		-e RACK_ENV=production \
		-p 4567:4567 \
		$(IMAGE)

# These tests require the dev-server or docker-dev-server above to be running
test:
	rspec

fetch-live-json-datafiles:
	pod=$$(kubectl -n how-out-of-date-are-we get pods -o name); \
		for file in $$(kubectl -n how-out-of-date-are-we exec $${pod} ls data); do \
		  echo $${file}; \
			kubectl -n how-out-of-date-are-we exec $${pod} cat data/$${file} > data/$${file}; \
		done

