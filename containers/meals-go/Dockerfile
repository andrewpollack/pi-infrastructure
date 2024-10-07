# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder
ARG TARGETPLATFORM

WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code and build the binary
COPY . .

RUN case "$TARGETPLATFORM" in \
        "linux/amd64") \
            echo "Building for x86_64"; \
            GOARCH=amd64 go build -o app .; \
            ;; \
        "linux/arm/v7") \
            echo "Building for ARMv7 (Raspberry Pi)"; \
            GOARCH=arm GOARM=7 go build -o app .; \
            ;; \
        "linux/arm64") \
            echo "Building for ARM64 (Raspberry Pi 64-bit)"; \
            GOARCH=arm64 go build -o app .; \
            ;; \
        *) echo "Unsupported architecture: $TARGETPLATFORM"; exit 1 ;; \
    esac

FROM alpine:3.20.3

WORKDIR /app

# Copy the compiled Go binary from the build stage
COPY --from=builder /app/app .

ENV SERVE_PORT=8000

# Run the Go binary
CMD ["./app"]
