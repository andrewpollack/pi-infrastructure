SHELL := /bin/bash

include variables.mk

setup:
	@./setup-node.sh

dev:
	@npm run dev

push:
	@echo "Building and pushing Docker image..."
	@docker buildx build --platform linux/amd64,linux/arm64 -t $(REPO_PATH) --push .
