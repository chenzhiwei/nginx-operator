QUAY_REPO ?= quay.io/siji
IMAGE_NAME ?= nginx
OPERATOR_NAME ?= nginx-operator
VERSION ?= 0.0.1

build:
	CGO_ENABLED=0 go build -o build/_output/bin/$(OPERATOR_NAME) cmd/manager/main.go
	@strip build/_output/bin/$(OPERATOR_NAME) || true

image:
	docker build -t $(QUAY_REPO)/$(IMAGE_NAME):$(VERSION) -f build/Dockerfile .

push-image:
	docker push $(QUAY_REPO)/$(IMAGE_NAME):$(VERSION)

push-csv:
	QUAY_REPO=$(QUAY_REPO) OPERATOR_NAME=$(OPERATOR_NAME) VERSION=$(VERSION) misc/push-csv.sh
