# ------------------------------------------------------------------------------
# Stage 1: Build the Go binary
# ------------------------------------------------------------------------------
# Force the builder platform to be linux/amd64 for the Go build, since we can
# target different architectures using GOOS/GOARCH while retaining non-emulator
# build speed.
ARG BUILDERPLATFORM=linux/amd64

FROM --platform=$BUILDERPLATFORM golang:1.21-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
# Handle ARM variant (v7) if specified
RUN if [ "$TARGETARCH" = "arm" ] && [ "$TARGETVARIANT" = "v7" ]; then \
      GOARM=7 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o flight-tracker; \
    else \
      GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o flight-tracker; \
    fi

# ------------------------------------------------------------------------------
# Stage 2: Build the final image for the target architecture
# ------------------------------------------------------------------------------
ARG TARGETPLATFORM
FROM --platform=$TARGETPLATFORM alpine:3.18

# Add CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -H -h /app appuser
WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/flight-tracker .

# Set ownership to non-root user
RUN chown -R appuser:appuser /app

# Use non-root user
USER appuser

# Expose the metrics port
EXPOSE 8080

# Set environment variables
ENV PORT=8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/metrics || exit 1

# Command to run the executable
CMD ["./flight-tracker"]
