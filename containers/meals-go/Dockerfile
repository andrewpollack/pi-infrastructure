# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder
ARG TARGETPLATFORM

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

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

# Stage 2: Build the Runtime environment, including the wkhtmltopdf binary for PDF generation
FROM debian:bullseye-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends wkhtmltopdf ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app .

CMD ["./app"]
