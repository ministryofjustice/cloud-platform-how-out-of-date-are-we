IMAGE := ministryofjustice/cloud-platform-how-out-of-date-are-we:2.3

.built-image: app.rb Gemfile* makefile views/*
	docker build -t $(IMAGE) .
	docker push $(IMAGE)
	touch .built-image

build: .built-image
	cd updater-image/; make build

run:
	docker-compose up --build

update:
	docker-compose build updater
	docker-compose run updater ./update.sh

dev-server:
	API_KEY=soopersekrit ./app.rb -o 0.0.0.0

# These tests require the dev-server above to be running
test:
	rspec
