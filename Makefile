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

generate-csv:
	operator-sdk generate csv --csv-version $(VERSION) --update-crds

push-csv:
	QUAY_REPO=$(QUAY_REPO) OPERATOR_NAME=$(OPERATOR_NAME) VERSION=$(VERSION) misc/push-csv.sh
