SHELL := /bin/bash

include variables.mk

setup:
	@./setup-node.sh

dev:
	@npm run dev

push:
	@docker buildx build --platform linux/amd64,linux/arm64 -t $(REPO_PATH) --push .
	@bazel run //containers/ops:update_pods -- $(POD_NAME) $(REGISTRY)

.PHONY: setup dev push
