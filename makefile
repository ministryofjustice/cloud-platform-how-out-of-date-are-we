IMAGE := ministryofjustice/cloud-platform-how-out-of-date-are-we:3.11
DEV_NAMESPACE := cloud-platform-reports-dev
PROD_NAMESPACE := cloud-platform-reports-prod
CRONJOB_NAMESPACE := concourse-main

deploy:
	make deploy-webapp
	make deploy-cronjobs

upgrade:
	make upgrade-webapp
	make upgrade-cronjobs

deploy-webapp:
	aws eks update-kubeconfig --name live \
	  && helm install \
			--generate-name \
			--namespace $(PROD_NAMESPACE) \
			./cloud-platform-reports

deploy-cronjobs:
	aws eks update-kubeconfig --name manager \
		&& helm install \
			--generate-name \
			--namespace $(CRONJOB_NAMESPACE) \
			./cloud-platform-reports-cronjobs \
			--values cloud-platform-reports-cronjobs/secrets.yaml

upgrade-webapp:
	aws eks update-kubeconfig --name live \
		&& helm upgrade \
			$$(helm ls --short --namespace $(PROD_NAMESPACE) | grep cloud-platform-reports) \
			--namespace $(PROD_NAMESPACE) \
			./cloud-platform-reports

upgrade-cronjobs:
	aws eks update-kubeconfig --name manager \
		&& helm upgrade \
			$$(helm ls --short --namespace $(CRONJOB_NAMESPACE) | grep cloud-platform-reports-cronjobs) \
			--namespace $(CRONJOB_NAMESPACE) \
			./cloud-platform-reports-cronjobs \
			--values cloud-platform-reports-cronjobs/secrets.yaml

dev-deploy:
	make dev-deploy-webapp
	make dev-deploy-cronjobs

dev-upgrade:
	make dev-upgrade-webapp
	make dev-upgrade-cronjobs

dev-deploy-webapp:
	aws eks update-kubeconfig --name live \
	  && helm install \
			--generate-name \
			--namespace $(DEV_NAMESPACE) \
			./cloud-platform-reports \
			--values cloud-platform-reports/secrets.yaml \
			--values cloud-platform-reports/values-dev.yaml

dev-deploy-cronjobs:
	aws eks update-kubeconfig --name manager \
		&& helm install \
			--generate-name \
			--namespace $(CRONJOB_NAMESPACE) \
			./cloud-platform-reports-cronjobs \
			--values cloud-platform-reports-cronjobs/values-dev.yaml \
			--values cloud-platform-reports-cronjobs/secrets.yaml

dev-upgrade-webapp:
	aws eks update-kubeconfig --name live \
		&& helm upgrade \
			$$(helm ls --short --namespace $(DEV_NAMESPACE) | grep cloud-platform-reports) \
			--namespace $(DEV_NAMESPACE) \
			./cloud-platform-reports \
			--values cloud-platform-reports/secrets.yaml \
			--values cloud-platform-reports/values-dev.yaml

dev-upgrade-cronjobs:
	aws eks update-kubeconfig --name manager \
		&& helm upgrade \
			$$(helm ls --short --namespace $(CRONJOB_NAMESPACE) | grep cloud-platform-reports-cronjobs) \
			--namespace $(CRONJOB_NAMESPACE) \
			./cloud-platform-reports-cronjobs \
			--values cloud-platform-reports-cronjobs/values-dev.yaml \
			--values cloud-platform-reports-cronjobs/secrets-dev.yaml

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


dev-uninstall:
	kubectl config use-context live-1 \
	  && helm uninstall --namespace $(DEV_NAMESPACE) $$(helm ls --short --namespace $(DEV_NAMESPACE))


dev-uninstall-cronjobs:
	kubectl config use-context manager \
		&& helm uninstall --namespace $(CRONJOB_NAMESPACE) $$(helm ls --short --namespace $(CRONJOB_NAMESPACE))

		./cloud-platform-reports \
		--values cloud-platform-reports/secrets.yaml \
		--values cloud-platform-reports/values-dev.yaml


