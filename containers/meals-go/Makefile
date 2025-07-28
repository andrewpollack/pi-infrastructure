SHELL := /bin/bash

include variables.mk

push:
	@bazel run //containers/meals-go:push
	@../ops/update-pods.sh $(POD_NAME) $(REGISTRY)

local:
	@bazel run //containers/meals-go:load
	@docker compose up

test-all:
	@bazel test //containers/meals-go/...

.PHONY: push local test-all
