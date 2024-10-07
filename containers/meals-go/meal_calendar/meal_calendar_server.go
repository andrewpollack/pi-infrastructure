package meal_calendar

import (
	"fmt"
	"meals/calendar"
	"meals/meal_collection"
	"net/http"
	"os"
	"time"
)

// helloHandler is the function that handles HTTP requests and responds with "Hello"
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Set the content type to text/html to let the browser know it's an HTML response
	w.Header().Set("Content-Type", "text/html")

	collection, _ := meal_collection.ReadMealCollection("")
	currYear, currMonth, _ := time.Now().Date()
	var nextMonth time.Month
	var nextYear int
	if currMonth == time.December {
		nextMonth = time.January
		nextYear = currYear + 1
	} else {
		nextMonth = currMonth + 1
		nextYear = currYear
	}

	currMonthMealCalendar := NewCalendar(*calendar.NewCalendar(currYear, currMonth), collection)
	nextMonthMealCalendar := NewCalendar(*calendar.NewCalendar(nextYear, nextMonth), collection)

	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Meals</title>
			<style>
				table, th, td {
					border: 1px solid black;
				}
			</style>
		</head>
		<body>
			%s

			%s
		</body>
		</html>
	`, currMonthMealCalendar.RenderHTMLCalendar(), nextMonthMealCalendar.RenderHTMLCalendar())
}

func RunServer() {
	// Route the root URL ("/") to the helloHandler
	http.HandleFunc("/", helloHandler)

	// Start the server on port 8080 and log any errors
	port := os.Getenv("SERVE_PORT")
	fmt.Printf("Starting server on :%s...", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
