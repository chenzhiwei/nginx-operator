QUAY_REPO ?= quay.io/siji
IMAGE_NAME ?= nginx
OPERATOR_NAME ?= nginx-operator
VERSION ?= 0.0.1

code:
	go mod tidy

build: code
	CGO_ENABLED=0 go build -o build/_output/bin/$(OPERATOR_NAME) cmd/manager/main.go
	@strip build/_output/bin/$(OPERATOR_NAME) || true

image: build
	docker build -t $(QUAY_REPO)/$(IMAGE_NAME):$(VERSION) -f build/Dockerfile .

push-image:
	docker push $(QUAY_REPO)/$(IMAGE_NAME):$(VERSION)

code-gen:
	@echo Updating the deep copy files with the changes in the API
	operator-sdk generate k8s
	@echo Updating the CRD files with the OpenAPI validations
	operator-sdk generate crds

generate-csv:
	@echo Updating/Generating a ClusterServiceVersion YAML manifest for the operator
	operator-sdk generate csv --csv-version $(VERSION) --update-crds

push-csv:
	QUAY_REPO=$(QUAY_REPO) OPERATOR_NAME=$(OPERATOR_NAME) VERSION=$(VERSION) misc/push-csv.sh
