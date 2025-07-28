SHELL := /bin/bash

include variables.mk

push:
	@bazel run //containers/flight-tracker-go:push
	@../ops/update-pods.sh $(POD_NAME) $(REGISTRY)

test-all:
	@bazel test //containers/flight-tracker-go/...

.PHONY: push test-all
