package main

import (
	"flag"
	"log"
	"meals/meal_backend"
	"meals/meal_calendar"
	"meals/meal_db_sync"
	"meals/meal_email"
	"os"
	"strings"
)

type Config struct {
	// Application configuration
	RunMode            string
	PostgresURL        string
	BucketName         string
	BucketKey          string
	IgnoreCutoff       bool
	DeploymentPassword string

	// CORS configuration
	AllowOrigins  []string
	DomainName    string
	JWTSigningKey []byte

	// DBSync configuration
	SyncCleanTable bool
	SyncLongLive   bool

	// Email configuration
	EmailSender    string
	EmailReceivers string
}

var (
	runMode            = flag.String("run_mode", envString("RUN_MODE", ""), "Application run mode: backend, email, db_sync, legacy")
	postgresURL        = flag.String("postgres_url", envString("POSTGRES_URL", ""), "Postgres URL for database connection")
	bucketName         = flag.String("bucket_name", envString("BUCKET_NAME", ""), "S3 bucket name for accessing meals data")
	bucketKey          = flag.String("bucket_key", envString("BUCKET_KEY", ""), "S3 bucket key for accessing meals data")
	ignoreCutoff       = flag.Bool("ignore_cutoff", envBool("IGNORE_CUTOFF", false), "Whether to ignore first-of-month meal cutoff")
	syncCleanTable     = flag.Bool("clean_table", envBool("CLEAN_TABLE", false), "Remove any unseen keys from database on sync")
	syncLongLive       = flag.Bool("long_live", envBool("LONG_LIVE", false), "Whether sync job should run indefinitely")
	emailSender        = flag.String("sender_email", envString("SENDER_EMAIL", ""), "Email address to send from")
	emailReceivers     = flag.String("receiver_emails", envString("RECEIVER_EMAILS", ""), "Comma-separated email addresses to send to")
	allowOrigins       = flag.String("allowed_origins", envString("ALLOWED_ORIGINS", "http://localhost:5173"), "Comma-separated list of allowed origins for CORS")
	domainName         = flag.String("domain_name", envString("DOMAIN_NAME", ""), "Domain name for the application")
	JWTSigningKey      = flag.String("jwt_signing_key", envString("JWT_SIGNING_KEY", "my-secret-key"), "JWT signing key for authentication")
	deploymentPassword = flag.String("deployment_password", envString("DEPLOYMENT_PASSWORD", "temp"), "Password for deployment")
)

func envString(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true"
	}
	return fallback
}

func parseFlags() Config {
	flag.Parse()

	// Parse comma separated allowed origins
	var allowOriginsSlice []string
	if *allowOrigins != "" {
		allowOriginsSlice = strings.Split(*allowOrigins, ",")
	}

	return Config{
		RunMode:            *runMode,
		PostgresURL:        *postgresURL,
		BucketName:         *bucketName,
		BucketKey:          *bucketKey,
		IgnoreCutoff:       *ignoreCutoff,
		SyncCleanTable:     *syncCleanTable,
		SyncLongLive:       *syncLongLive,
		EmailSender:        *emailSender,
		EmailReceivers:     *emailReceivers,
		DomainName:         *domainName,
		DeploymentPassword: *deploymentPassword,
		JWTSigningKey:      []byte(*JWTSigningKey),
		AllowOrigins:       allowOriginsSlice,
	}
}

func main() {
	c := parseFlags()

	switch c.RunMode {
	case "backend":
		mealBackendConfig := meal_backend.Config{
			PostgresURL:        c.PostgresURL,
			EmailSender:        c.EmailSender,
			EmailReceivers:     c.EmailReceivers,
			IgnoreCutoff:       c.IgnoreCutoff,
			AllowOrigins:       c.AllowOrigins,
			DomainName:         c.DomainName,
			JWTSigningKey:      c.JWTSigningKey,
			DeploymentPassword: c.DeploymentPassword,
		}

		mealBackendConfig.RunBackend()
	case "email":
		mealEmailConfig := meal_email.Config{
			PostgresURL:  c.PostgresURL,
			EmailService: meal_email.SES,
			Sender:       c.EmailSender,
			Receivers:    c.EmailReceivers,
			IgnoreCutoff: c.IgnoreCutoff,
		}

		err := mealEmailConfig.CreateAndSendEmail()
		if err != nil {
			log.Printf("Error: %s\n", err)
		}
	case "db_sync":
		mealDbSyncConfig := meal_db_sync.Config{
			PostgresURL: c.PostgresURL,
			BucketName:  c.BucketName,
			BucketKey:   c.BucketKey,
			CleanTable:  c.SyncCleanTable,
			LongLive:    c.SyncLongLive,
		}

		err := mealDbSyncConfig.SyncMealsWrapper()
		if err != nil {
			log.Printf("Error: %s\n", err)
		}
	case "legacy":
		// Legacy frontend+backend combined in one service
		mealsLegacyCalendarConfig := meal_calendar.Config{
			BucketName: c.BucketName,
			BucketKey:  c.BucketKey,
			Port:       8001,
		}

		mealsLegacyCalendarConfig.RunServer()
	default:
		log.Printf("Invalid RUN_MODE: %s\n", c.RunMode)
		os.Exit(1)
	}

	log.Println("Application finished.")
}
