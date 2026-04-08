IMAGE_NAME ?= score-argocd-cmp
IMAGE_TAG ?= latest
SCORE_K8S_VERSION ?= latest

.PHONY: build test go-test clean

build:
	docker build \
		--build-arg SCORE_K8S_VERSION=$(SCORE_K8S_VERSION) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) .

go-test:
	go test ./...

test: build go-test
	./tests/test-generate.sh $(IMAGE_NAME):$(IMAGE_TAG)

clean:
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true
