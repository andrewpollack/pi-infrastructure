package main

import (
	"fmt"
	"meals/meal_backend"
	"meals/meal_calendar"
	"meals/meal_db_sync"
	"meals/meal_email"
	"os"
)

func main() {
	runMode := os.Getenv("RUN_MODE")

	switch runMode {
	case "backend":
		meal_backend.RunBackend()
	case "email":
		useSES := true
		err := meal_email.CreateAndSendEmail(useSES)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	case "db_sync":
		err := meal_db_sync.SyncMeals()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	case "legacy":
		// Legacy frontend+backend combined in one service
		meal_calendar.RunServer()
	default:
		fmt.Printf("Invalid RUN_MODE: %s\n", runMode)
		os.Exit(1)
	}

	fmt.Println("Application finished.")
}
