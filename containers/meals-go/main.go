package main

import (
	"flag"
	"log"
	"meals/meal_backend"
	"meals/meal_calendar"
	"meals/meal_db_sync"
	"meals/meal_email"
	"os"
)

type Config struct {
	RunMode        string
	PostgresURL    string
	BucketName     string
	BucketKey      string
	SyncCleanTable bool
	SenderEmail    string
	HardcodedMeals []string
	ReceiverEmails string
	DryRun         bool
	LongLive       bool
	IgnoreCutoff   bool
}

func parseFlags() Config {
	runMode := flag.String("run_mode", os.Getenv("RUN_MODE"), "Application run mode: backend, email, db_sync, legacy")
	postgresURL := flag.String("postgres_url", os.Getenv("POSTGRES_URL"), "Postgres URL for database connection")
	bucketName := flag.String("bucket_name", os.Getenv("BUCKET_NAME"), "S3 bucket name for accessing meals data")
	bucketKey := flag.String("bucket_key", os.Getenv("BUCKET_KEY"), "S3 bucket key for accessing meals data")
	syncCleanTable := flag.Bool("clean_table", os.Getenv("CLEAN_TABLE") == "true", "Remove any unseen keys from database on sync")
	senderEmail := flag.String("sender_email", os.Getenv("SENDER_EMAIL"), "Email address to send from")
	receiverEmails := flag.String("receiver_emails", os.Getenv("RECEIVER_EMAILS"), "Comma-separated email addresses to send to")
	longLive := flag.Bool("long_live", os.Getenv("LONG_LIVE") == "true", "Whether sync job should run indefinitely")
	ignoreCutoff := flag.Bool("ignore_cutoff", os.Getenv("IGNORE_CUTOFF") == "true", "Whether to ignore first of month meal cutoff")
	hardcodedMeal1 := flag.String("h_1", os.Getenv("H_1"), "1 hardcoded meal")
	hardcodedMeal2 := flag.String("h_2", os.Getenv("H_2"), "2 hardcoded meal")
	hardcodedMeal3 := flag.String("h_3", os.Getenv("H_3"), "3 hardcoded meal")
	hardcodedMeal4 := flag.String("h_4", os.Getenv("H_4"), "4 hardcoded meal")
	hardcodedMeal5 := flag.String("h_5", os.Getenv("H_5"), "5 hardcoded meal")
	hardcodedMeal6 := flag.String("h_6", os.Getenv("H_6"), "6 hardcoded meal")
	hardcodedMeal7 := flag.String("h_7", os.Getenv("H_7"), "7 hardcoded meal")
	dryRun := flag.Bool("dry_run", os.Getenv("DRY_RUN") == "true", "Dry run mode for testing email")
	flag.Parse()

	// Check if hardcoded meals are set
	var hardcodedMeals []string
	if *hardcodedMeal1 != "" && *hardcodedMeal2 != "" && *hardcodedMeal3 != "" && *hardcodedMeal4 != "" && *hardcodedMeal5 != "" && *hardcodedMeal6 != "" && *hardcodedMeal7 != "" {
		hardcodedMeals = []string{*hardcodedMeal1, *hardcodedMeal2, *hardcodedMeal3, *hardcodedMeal4, *hardcodedMeal5, *hardcodedMeal6, *hardcodedMeal7}
	}

	return Config{
		RunMode:        *runMode,
		PostgresURL:    *postgresURL,
		BucketName:     *bucketName,
		BucketKey:      *bucketKey,
		SyncCleanTable: *syncCleanTable,
		SenderEmail:    *senderEmail,
		ReceiverEmails: *receiverEmails,
		DryRun:         *dryRun,
		HardcodedMeals: hardcodedMeals,
		LongLive:       *longLive,
		IgnoreCutoff:   *ignoreCutoff,
	}
}

func main() {
	c := parseFlags()

	switch c.RunMode {
	case "backend":
		mealBackendConfig := meal_backend.Config{
			PostgresURL:    c.PostgresURL,
			SenderEmail:    c.SenderEmail,
			ReceiverEmails: c.ReceiverEmails,
			IgnoreCutoff:   c.IgnoreCutoff,
		}

		mealBackendConfig.RunBackend()
	case "email":
		mealEmailConfig := meal_email.Config{
			PostgresURL:    c.PostgresURL,
			UseSES:         true,
			SenderEmail:    c.SenderEmail,
			ReceiverEmails: c.ReceiverEmails,
			HardcodedMeals: c.HardcodedMeals,
			DryRun:         c.DryRun,
			IgnoreCutoff:   c.IgnoreCutoff,
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
			LongLive:    c.LongLive,
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
