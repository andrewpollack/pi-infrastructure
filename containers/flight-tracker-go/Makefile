SHELL := /bin/bash

include variables.mk

push:
	@docker buildx build --platform linux/amd64,linux/arm64 -t $(REPO_PATH) --push .
	@../ops/update-pods.sh $(POD_NAME) $(REGISTRY)

test-all:
	go test ./...

.PHONY: push test-all
