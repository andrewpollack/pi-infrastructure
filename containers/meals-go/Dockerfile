# ------------------------------------------------------------------------------
# Stage 1: Build the Go binary, using an AMD64 container
# ------------------------------------------------------------------------------
# Force the builder platform to be linux/amd64 for the Go build, since we can
# target different architectures using GOOS/GOARCH while retaining non-emulator
# build speed.
ARG BUILDERPLATFORM=linux/amd64

FROM --platform=$BUILDERPLATFORM golang:1.24-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app .

# ------------------------------------------------------------------------------
# Stage 2: Build the final image for the target architecture
# ------------------------------------------------------------------------------
FROM debian:bullseye-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends wkhtmltopdf ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

# Copy the built binary and the migrations directory from the builder stage
COPY --from=builder /app/app .
COPY --from=builder /app/migrations ./migrations

CMD ["./app"]
