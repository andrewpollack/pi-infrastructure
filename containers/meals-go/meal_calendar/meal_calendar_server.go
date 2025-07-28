package meal_calendar

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/andrewpollack/pi-infrastructure/containers/meals-go/calendar"
	"github.com/andrewpollack/pi-infrastructure/containers/meals-go/meal_collection"
)

type Config struct {
	BucketName string
	BucketKey  string
	Port       int
}

func (c Config) mealCalendarHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch meal data from S3
	mealData, err := meal_collection.OpenFromS3(c.BucketName, c.BucketKey)
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	// Return HTML
	w.Header().Set("Content-Type", "text/html")

	// Decode the meal collection
	mealCollection, _ := meal_collection.ReadMealCollectionFromReader(mealData)

	// Get current date info
	currYear, currMonth, _ := time.Now().Date()

	var nextMonth time.Month
	var nextYear int

	// Determine the next month/year
	if currMonth == time.December {
		nextMonth = time.January
		nextYear = currYear + 1
	} else {
		nextMonth = currMonth + 1
		nextYear = currYear
	}

	// Build two calendars: current month + next month
	currMonthMealCalendar := NewCalendar(*calendar.NewCalendar(currYear, currMonth), mealCollection)
	nextMonthMealCalendar := NewCalendar(*calendar.NewCalendar(nextYear, nextMonth), mealCollection)

	// Build the HTML list of all items
	endList := "<h2>ALL ITEMS</h2>\n\n<ul>\n"
	for _, item := range mealCollection {
		itemName := item.Name
		if len(item.Ingredients) == 0 &&
			item.Name != "Leftovers" &&
			item.Name != "Out" {
			itemName += "*"
		}

		endList += "\t<li>"
		if item.URL != nil {
			endList += fmt.Sprintf("<a href=\"%s\">%s</a>", *item.URL, itemName)
		} else {
			endList += itemName
		}
		endList += "</li>\n"
	}
	endList += "</ul>"

	// Render final HTML
	_, err = fmt.Fprintf(w, `<!DOCTYPE html>
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
</html>`,
		currMonthMealCalendar.RenderHTMLCalendar(),
		nextMonthMealCalendar.RenderHTMLCalendar(),
		endList,
	)

	if err != nil {
		log.Println("Error writing response:", err)
		return
	}
}

func (c Config) RunServer() {
	http.HandleFunc("/", c.mealCalendarHandler)

	log.Printf("Starting server on :%d...", c.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil)
	if err != nil {
		log.Println("Error starting server:", err)
	}
}
