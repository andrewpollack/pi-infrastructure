# Flight Tracker Go

## Overview
A Go-based flight tracking application that retrieves flight information for a specific tail number and exposes Prometheus metrics for monitoring.

## Features
- Fetch flight details from FlightAware
- Extract departure, arrival, and status information
- Prometheus metrics integration for monitoring
- Multi-architecture support (ARM/AMD)
- Containerized application for k3s deployment

## Prerequisites
- Go 1.21+
- Docker (for building container images)
- Prometheus (for metrics scraping)

## Prometheus Metrics
The application exposes the following metrics at the `/metrics` endpoint:

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|--------|
| `flight_tracker_requests_total` | Counter | Total number of flight information requests | - |
| `flight_tracker_errors_total` | Counter | Total number of flight information request errors | - |
| `flight_tracker_request_duration_seconds` | Histogram | Duration of flight information requests in seconds | - |
| `flight_tracker_status` | Gauge | Current flight status (1=in air, 0=landed/not flying) | `tail_number`, `owner` |
| `flight_tracker_distance_miles` | Gauge | Flight distance in miles | `tail_number` |
| `flight_tracker_fuel_gallons` | Gauge | Estimated fuel usage in gallons | `tail_number` |

## Usage
### Local Run
```bash
go mod download
go run main.go
```

The server will start on port 8080 with the following endpoints:
- `/flight?tail=<tail_number>` - Get flight information for a specific tail number
- `/metrics` - Prometheus metrics endpoint

### Docker Build
```bash
# Build for multiple architectures
docker buildx build --platform linux/amd64,linux/arm/v7 -t flight-tracker .

# Run the container
docker run -p 8080:8080 flight-tracker
```

### Kubernetes Deployment
The application is designed to be deployed in a k3s cluster. Example deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flight-tracker
spec:
  replicas: 1
  selector:
    matchLabels:
      app: flight-tracker
  template:
    metadata:
      labels:
        app: flight-tracker
    spec:
      containers:
      - name: flight-tracker
        image: flight-tracker:latest
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: flight-tracker
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
spec:
  selector:
    app: flight-tracker
  ports:
  - port: 8080
    targetPort: 8080
  type: NodePort
```

## Prometheus Integration
To configure Prometheus to scrape metrics from this application, add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'flight-tracker'
    scrape_interval: 15s
    static_configs:
      - targets: ['flight-tracker:8080']
```

## Limitations
- Depends on FlightAware website structure (may break if site changes)
- Requires active internet connection

## TODO
- Implement more robust error handling
- Add support for multiple tail numbers
- Create comprehensive unit tests
- Add caching mechanism
- Implement retry logic for API calls
