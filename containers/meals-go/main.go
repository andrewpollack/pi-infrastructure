package main

import (
	"fmt"
	"log"
	"meals/meal_calendar"
	"meals/meal_email"
	"os"
)

func main() {
	justEmail := os.Getenv("JUST_EMAIL")

	switch justEmail {
	case "false", "":
		meal_calendar.RunServer()
	default:
		srv, err := meal_email.AuthenticateGmail()
		if err != nil {
			log.Fatalf("Failed to authenticate with Gmail: %s", err.Error())
		}

		meal_email.CreateAndSendEmail(srv)
	}

	fmt.Println("Application finished.")
}
