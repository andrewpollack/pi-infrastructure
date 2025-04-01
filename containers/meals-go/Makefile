SHELL := /bin/bash

include variables.mk

push:
	@docker buildx build --platform linux/amd64,linux/arm64 -t $(REPO_PATH) --push .
	@../ops/update-pods.sh $(POD_NAME) $(REGISTRY)

run_image:
	docker run --pull always --env-file local.env $(REPO_PATH)

test-all:
	go test ./...

.PHONY: push run_image test-all
