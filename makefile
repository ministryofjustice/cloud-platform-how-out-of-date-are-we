IMAGE := ministryofjustice/cloud-platform-how-out-of-date-are-we:0.1

build: .built-image

.built-image: app.rb Gemfile* makefile views/*
	docker build -t $(IMAGE) .
	docker push $(IMAGE)
	touch .built-image

run: build
	docker run --rm \
		-p 4567:4567 \
		-e API_KEY=soopersekrit \
		-e RACK_ENV=production \
		-it $(IMAGE)

# Ensure you have a data/helm-whatup.json file
# NB: you will have to restart this process when you
# change files. Most changes will not be picked up by
# just reloading the web page.
dev-server:
	./app.rb -o 0.0.0.0
