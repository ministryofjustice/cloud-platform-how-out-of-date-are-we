IMAGE := ministryofjustice/cloud-platform-how-out-of-date-are-we-updater:1.3

.built-image: Dockerfile makefile update.sh module-versions.rb
	docker build -t $(IMAGE) .
	docker push $(IMAGE)
	touch .built-image

build: .built-image
