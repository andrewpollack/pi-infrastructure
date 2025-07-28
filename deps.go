// Package main is a dependency resolution helper for the Bazel build system.
// This file ensures that all required dependencies are properly resolved.
// The main function is a placeholder and is not intended to be executed.
package main

import (
	// Import to make prometheus client a direct dependency
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"

	// Import to make go-rpio a direct dependency
	_ "github.com/stianeikeland/go-rpio/v4"

	// Import meals-go dependencies
	_ "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	_ "github.com/aws/aws-sdk-go-v2"
	_ "github.com/aws/aws-sdk-go-v2/config"
	_ "github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/aws/aws-sdk-go-v2/service/ses"
	_ "github.com/bazelbuild/rules_go/go/runfiles"
	_ "github.com/gin-contrib/cors"
	_ "github.com/gin-gonic/gin"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/knadh/koanf/parsers/yaml"
	_ "github.com/knadh/koanf/providers/file"
	_ "github.com/knadh/koanf/v2"
	_ "github.com/lib/pq"
	_ "golang.org/x/exp/slices"
)

// This function exists only to establish the dependency chain
// It is not intended to be called or used in any way
func main() {
	// This is intentionally empty - it's just for dependency resolution
}
