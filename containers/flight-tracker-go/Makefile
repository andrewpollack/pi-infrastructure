SHELL := /bin/bash

include variables.mk

push:
	@bazel run //containers/flight-tracker-go:push
	@bazel run //containers/ops:update_pods -- $(POD_NAME) $(REGISTRY)

test-all:
	@bazel test //containers/flight-tracker-go/...

.PHONY: push test-all
