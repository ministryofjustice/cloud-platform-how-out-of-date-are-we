IMAGE := ministryofjustice/cloud-platform-how-out-of-date-are-we-updater:1.0

.built-image: Dockerfile makefile update.sh module-versions.rb
	docker build -t $(IMAGE) .
	docker push $(IMAGE)
	touch .built-image

# This is only required during development. Normally, the concourse
# pipeline does this.
run: .built-image
	docker run \
		-e AWS_ACCESS_KEY_ID=$${AWS_ACCESS_KEY_ID} \
		-e AWS_SECRET_ACCESS_KEY=$${AWS_SECRET_ACCESS_KEY} \
		-e KUBECONFIG_S3_BUCKET=$${KUBECONFIG_S3_BUCKET} \
		-e KUBECONFIG_S3_KEY=$${KUBECONFIG_S3_KEY} \
		-e KUBECONFIG=$${KUBECONFIG} \
		-e KUBE_CLUSTER=$${KUBE_CLUSTER} \
		-e HTTP_ENDPOINT=$${HTTP_ENDPOINT} \
	  --rm -it $(IMAGE) \
		./update.sh
