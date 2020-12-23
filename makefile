IMAGE := ministryofjustice/cloud-platform-how-out-of-date-are-we:2.11
DEV_NAMESPACE := cloud-platform-reports-dev
CRONJOB_NAMESPACE := concourse-main

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

# This requires that the following ENV vars are set:
#   * DYNAMODB_TABLE_NAME
#   * DYNAMODB_ACCESS_KEY_ID
#   * DYNAMODB_SECRET_ACCESS_KEY
fetch-live-json-datafiles:
	mkdir -p data/namespace/costs
	./fetch-data-from-dynamodb.rb

dev-deploy:
	kubectl config use-context live-1 \
	  && helm install \
			--generate-name \
			--namespace $(DEV_NAMESPACE) \
			./cloud-platform-reports \
			--values cloud-platform-reports/secrets.yaml \
			--values cloud-platform-reports/values-dev.yaml

dev-uninstall:
	kubectl config use-context live-1 \
	  && helm uninstall --namespace $(DEV_NAMESPACE) $$(helm ls --short --namespace $(DEV_NAMESPACE))

dev-deploy-cronjobs:
	kubectl config use-context manager \
		&& helm install \
			--generate-name \
			--namespace $(CRONJOB_NAMESPACE) \
			./cloud-platform-reports-cronjobs \
			--values cloud-platform-reports/values-dev.yaml \
			--values cloud-platform-reports-cronjobs/values-dev.yaml \
			--values cloud-platform-reports-cronjobs/secrets.yaml

dev-uninstall-cronjobs:
	kubectl config use-context manager \
		&& helm uninstall --namespace $(CRONJOB_NAMESPACE) $$(helm ls --short --namespace $(CRONJOB_NAMESPACE))

		./cloud-platform-reports \
		--values cloud-platform-reports/secrets.yaml \
		--values cloud-platform-reports/values-dev.yaml


