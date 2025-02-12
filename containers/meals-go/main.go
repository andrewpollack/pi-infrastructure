package main

import (
	"fmt"
	"log"
	"meals/meal_backend"
	"meals/meal_calendar"
	"meals/meal_email"
	"os"
)

func main() {
	runMode := os.Getenv("RUN_MODE")

	switch runMode {
	case "server":
		meal_calendar.RunServer()
	case "backend", "":
		meal_backend.RunBackend()
	case "email":
		srv, err := meal_email.AuthenticateGmail()
		if err != nil {
			log.Fatalf("Failed to authenticate with Gmail: %s", err.Error())
		}

		meal_email.CreateAndSendEmail(srv)
	}

	fmt.Println("Application finished.")
}
