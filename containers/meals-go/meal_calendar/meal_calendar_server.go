package meal_calendar

import (
	"fmt"
	"meals/calendar"
	"meals/meal_collection"
	"net/http"
	"os"
	"sort"
	"strings"
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

	var flattenedItems []meal_collection.Item
	for _, item := range collection {
		flattenedItems = append(flattenedItems, item.Items...)
	}

	sort.Slice(flattenedItems, func(i, j int) bool {
		return strings.ToLower(flattenedItems[i].Name) < strings.ToLower(flattenedItems[j].Name)
	})

	endList := "<h2> ALL ITEMS </h2>\n\n<ul>\n"
	for _, item := range flattenedItems {
		itemName := item.Name
		if len(item.Ingredients) == 0 {
			if item.Name != "LEFTOVERS" && item.Name != "OUT" {
				itemName += "*"
			}
		}

		endList += "\t<li>"
		if item.URL != nil {
			endList += fmt.Sprintf("<a href=\"%s\">%s</a>", *item.URL, itemName)
		} else {
			endList += itemName
		}
		endList += "</li>\n"
	}
	endList += "\n</ul>"

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

			%s
		</body>
		</html>
	`, currMonthMealCalendar.RenderHTMLCalendar(), nextMonthMealCalendar.RenderHTMLCalendar(), endList)
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
