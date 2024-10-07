package main

import (
	"log"
	"meals/meal_calendar"
	"meals/meal_email"
	"os"
)

func main() {
	if os.Getenv("JUST_EMAIL") == "false" {
		meal_calendar.RunServer()
	} else {
		srv, err := meal_email.AuthenticateGmail()
		if err != nil {
			log.Fatalf("Failed to authenticate with Gmail: %s", err.Error())
		}

		meal_email.CreateAndSendEmail(srv)
	}
}
