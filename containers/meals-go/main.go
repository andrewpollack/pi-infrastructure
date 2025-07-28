package main

import (
	"flag"
	"log"
	"os"

	"github.com/andrewpollack/pi-infrastructure/containers/meals-go/config"
	"github.com/andrewpollack/pi-infrastructure/containers/meals-go/meal_backend"
	"github.com/andrewpollack/pi-infrastructure/containers/meals-go/meal_calendar"
	"github.com/andrewpollack/pi-infrastructure/containers/meals-go/meal_db_sync"
	"github.com/andrewpollack/pi-infrastructure/containers/meals-go/meal_email"
)

type Config struct {
	// Application configuration
	RunMode            string
	DeploymentPassword string

	// CORS configuration
	JWTSigningKey []byte

	// DBSync configuration
	SyncCleanTable bool
	SyncLongLive   bool
}

var (
	conf               = flag.String("conf", envString("CONF", "/app/conf.yaml"), "Path to the config file")
	runMode            = flag.String("run_mode", envString("RUN_MODE", ""), "Application run mode: backend, email, db_sync, legacy")
	syncCleanTable     = flag.Bool("clean_table", envBool("CLEAN_TABLE", false), "Remove any unseen keys from database on sync")
	syncLongLive       = flag.Bool("long_live", envBool("LONG_LIVE", false), "Whether sync job should run indefinitely")
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

	return Config{
		RunMode:            *runMode,
		SyncCleanTable:     *syncCleanTable,
		SyncLongLive:       *syncLongLive,
		DeploymentPassword: *deploymentPassword,
		JWTSigningKey:      []byte(*JWTSigningKey),
	}
}

func main() {
	c := parseFlags()

	config.Load(*conf)

	switch c.RunMode {
	case "backend":
		mealBackendConfig := meal_backend.Config{
			PostgresURL:          config.Cfg.Database.Postgres.URL,
			PostgresMigrationDir: "file://migrations",
			EmailSender:          config.Cfg.Email.Sender,
			EmailReceivers:       config.Cfg.Email.Receivers,
			AllowOrigins:         config.Cfg.Server.AllowedOrigins,
			JWTSigningKey:        c.JWTSigningKey,
			DeploymentPassword:   c.DeploymentPassword,
		}

		mealBackendConfig.RunBackend()
	case "email":
		mealEmailConfig := meal_email.Config{
			PostgresURL:  config.Cfg.Database.Postgres.URL,
			EmailService: meal_email.SES,
			Sender:       config.Cfg.Email.Sender,
			Receivers:    config.Cfg.Email.Receivers,
		}

		err := mealEmailConfig.CreateAndSendEmail()
		if err != nil {
			log.Printf("Error: %s\n", err)
		}
	case "db_sync":
		mealDbSyncConfig := meal_db_sync.Config{
			PostgresURL: config.Cfg.Database.Postgres.URL,
			BucketName:  config.Cfg.AWS.Bucket.Name,
			BucketKey:   config.Cfg.AWS.Bucket.Key,
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
			BucketName: config.Cfg.AWS.Bucket.Name,
			BucketKey:  config.Cfg.AWS.Bucket.Key,
			Port:       8001,
		}

		mealsLegacyCalendarConfig.RunServer()
	default:
		log.Printf("Invalid RUN_MODE: %s\n", c.RunMode)
		os.Exit(1)
	}

	log.Println("Application finished.")
}
