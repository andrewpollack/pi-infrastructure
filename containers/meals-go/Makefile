# Makefile

# Include external variables file
include variables.mk

# Target to build and push the Docker image
build_and_push_image:
	@echo "Building and pushing Docker image..."
	docker buildx build --platform linux/amd64,linux/arm64 -t $(REPO_PATH) --push .

run_image:
	docker run --pull always --env-file local.env $(REPO_PATH)

test-all:
	go test ./...

.PHONY: deploy_meals_image copy_files build_and_push_image
