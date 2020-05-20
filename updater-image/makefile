IMAGE := ministryofjustice/cloud-platform-how-out-of-date-are-we-updater:2.3

.built-image: Dockerfile makefile update.sh module-versions.rb documentation-pages-to-review.rb helm-releases.rb
	docker build -t $(IMAGE) .
	docker push $(IMAGE)
	touch .built-image

build: .built-image
