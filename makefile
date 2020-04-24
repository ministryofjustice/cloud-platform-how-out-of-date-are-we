IMAGE := ministryofjustice/cloud-platform-how-out-of-date-are-we:1.4

.built-image: app.rb Gemfile* makefile views/*
	docker build -t $(IMAGE) .
	docker push $(IMAGE)
	touch .built-image

build: .built-image
	cd updater-image/; make build

run:
	docker-compose up --build

updater:
	docker-compose run updater ./update.sh

# Ensure you have a data/helm-whatup.json file
# NB: you will have to restart this process when you
# change files. Most changes will not be picked up by
# just reloading the web page.
dev-server:
	API_KEY=soopersekrit ./app.rb -o 0.0.0.0

# These tests require the dev-server above to be running
test:
	rspec
