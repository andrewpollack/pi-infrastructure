package main

import (
	"fmt"
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
		useSES := true
		err := meal_email.CreateAndSendEmail(useSES)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}

	fmt.Println("Application finished.")
}
