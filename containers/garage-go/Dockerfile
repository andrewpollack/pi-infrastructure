# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code and build the binary
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -ldflags="-s -w" \
    -o doorcheck \
    .

FROM alpine:3.21.0
RUN mkdir /app && adduser -h /app -D doorcheck
WORKDIR /app

# Copy the compiled Go binary from the build stage
COPY --chown=doorcheck --from=builder /app/doorcheck .

ENV SERVE_PORT=2135

# Run the Go binary
CMD ["/app/doorcheck"]
