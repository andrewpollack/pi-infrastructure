package meal_email

import (
	"fmt"
	"log"
	"meals/calendar"
	"meals/meal_collection"
	"os"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

type Date struct {
	Year  int
	Month int
	Day   int
}

func GetDaysOfCurrentWeek(date Date) []Date {
	currMonthCalendar := calendar.NewCalendar(date.Year, time.Month(date.Month))
	var nextMonthCalendar *calendar.Calendar
	if date.Month == 12 {
		nextMonthCalendar = calendar.NewCalendar(date.Year+1, 1)
	} else {
		nextMonthCalendar = calendar.NewCalendar(date.Year, time.Month(date.Month+1))
	}

	dayIndex := currMonthCalendar.GetWeekIndexOfDay(date.Day)

	var fullWeek []Date
	currCalendarWeek := currMonthCalendar.Weeks[dayIndex]
	nextMonthStartWeek := nextMonthCalendar.Weeks[0]
	for i := 0; i < 7; i++ {
		if currCalendarWeek[i].Number == 0 {
			// Use next month instead.
			fullWeek = append(fullWeek, Date{
				Year:  nextMonthCalendar.Year,
				Month: int(nextMonthCalendar.Month),
				Day:   nextMonthStartWeek[i].Number,
			})
		} else {
			fullWeek = append(fullWeek, Date{
				Year:  currMonthCalendar.Year,
				Month: int(currMonthCalendar.Month),
				Day:   currCalendarWeek[i].Number,
			})
		}
	}

	return fullWeek
}

func GetDaysOfNextWeek(date Date) []Date {
	currMonthCalendar := calendar.NewCalendar(date.Year, time.Month(date.Month))
	daysInMonth := currMonthCalendar.DaysInMonth()

	nextWeekYear := date.Year
	nextWeekMonth := date.Month
	nextWeekDay := date.Day + 7

	if nextWeekDay > daysInMonth {
		if nextWeekMonth == 12 {
			nextWeekMonth = 1
			nextWeekYear += 1
		} else {
			nextWeekMonth += 1
		}
		nextWeekDay -= daysInMonth
	}

	return GetDaysOfCurrentWeek(Date{
		Year:  nextWeekYear,
		Month: nextWeekMonth,
		Day:   nextWeekDay,
	})
}

type YearMonth struct {
	Year  int
	Month int
}

func CreateGroceryEmailMessage(meals []meal_collection.Meal) string {
	var sb strings.Builder

	groceryCollection := meal_collection.MealsToGroceryItems(meals)

	for _, aisle := range meal_collection.AllAisles {
		itemsForAisle := groceryCollection[aisle]

		// Sort by Name
		sort.Slice(itemsForAisle, func(i, j int) bool {
			return itemsForAisle[i].Name < itemsForAisle[j].Name
		})

		fmt.Fprintf(&sb, "<h4>%s</h4>\n", aisle)

		// If no items for this aisle, show "NONE"
		if len(itemsForAisle) == 0 {
			sb.WriteString("<p>NONE</p>\n<br>\n")
			continue
		}

		// Otherwise, create a UL of items
		sb.WriteString("<ul style='margin-left: 20px;'>\n")
		for _, gi := range itemsForAisle {
			fmt.Fprintf(&sb, "<li>%s</li>\n", gi.String())
		}
		sb.WriteString("</ul>\n<br>\n")
	}

	return sb.String()
}

func useHardcodedValues(collection meal_collection.MealCollection) []meal_collection.Meal {
	var flattenedItems []meal_collection.Meal
	for _, item := range collection {
		flattenedItems = append(flattenedItems, item.Items...)
	}

	arr := [7]string{
		os.Getenv("H_1"),
		os.Getenv("H_2"),
		os.Getenv("H_3"),
		os.Getenv("H_4"),
		os.Getenv("H_5"),
		os.Getenv("H_6"),
		os.Getenv("H_7"),
	}

	var allItems []meal_collection.Meal
	for i, v := range arr {
		if i == 4 {
			allItems = append(allItems, meal_collection.Meal{
				Name: "LEFTOVERS",
			})
			continue
		}
		if i == 5 {
			allItems = append(allItems, meal_collection.Meal{
				Name: "OUT",
			})
			continue
		}

		for _, fullItem := range flattenedItems {
			if fullItem.Name == v {
				allItems = append(allItems, fullItem)
				break
			}
		}
	}

	return allItems
}

func GenerateEmailForNextWeek(date Date, collection meal_collection.MealCollection) string {
	daysOfWeek := GetDaysOfNextWeek(date)

	calendars := make(map[YearMonth][]meal_collection.Meal)

	var allItems []meal_collection.Meal
	switch os.Getenv("USE_HARDCODE") {
	case "false", "":
		for _, day := range daysOfWeek {
			currYearMonth := YearMonth{day.Year, day.Month}

			if _, exists := calendars[currYearMonth]; !exists {
				calendars[currYearMonth] = collection.GenerateMealsWholeYearNoCategories(*calendar.NewCalendar(day.Year, time.Month(day.Month)))
			}

			allItems = append(allItems, calendars[currYearMonth][day.Day-1])
		}
	default:
		allItems = useHardcodedValues(collection)
	}

	htmlBody := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Email</title>
	</head>
	<body>
	<h3>Meals:</h3>
`
	htmlBody += "<table border='1'>\n"
	htmlBody += "<thead>\n<tr>"

	fullDaysOfWeek := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	for _, day := range fullDaysOfWeek {
		htmlBody += fmt.Sprintf("<th>%s</th>\n", day)
	}

	htmlBody += "</tr>\n</thead>\n"
	htmlBody += "<tbody>\n<tr>"

	for i := range fullDaysOfWeek {
		htmlBody += fmt.Sprintf("<td>%s</td>\n", allItems[i].Name)
	}

	htmlBody += "</tr>\n</tbody>\n"
	htmlBody += "</table>\n\n"

	htmlBody += CreateGroceryEmailMessage(allItems)

	htmlBody += `
	</body>
	</html>
	`

	return htmlBody
}

func GenerateHeaderForNextWeek(date Date) string {
	daysOfWeek := GetDaysOfNextWeek(date)
	first := daysOfWeek[0]
	last := daysOfWeek[6]

	return fmt.Sprintf("Meals for %s %d -> %s %d ", time.Month(first.Month), first.Day, time.Month(last.Month), last.Day)
}

func CreateAndSendEmail(srv *gmail.Service) {
	currentTime := time.Now()

	// Extract the year, month, and day
	year := currentTime.Year()
	month := currentTime.Month()
	day := currentTime.Day()

	mealData, err := meal_collection.OpenFromS3()
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	collection, err := meal_collection.ReadMealCollection(mealData)
	if err != nil {
		fmt.Printf("Something went wrong reading meals: %s\n", err)

		return
	}

	currDate := Date{Year: year, Month: int(month), Day: day}
	from := os.Getenv("SENDER_EMAIL")
	to := os.Getenv("RECEIVER_EMAILS")
	subject := GenerateHeaderForNextWeek(currDate)
	body := GenerateEmailForNextWeek(currDate, collection)

	dryRun := os.Getenv("DRY_RUN")
	if dryRun == "true" {
		fmt.Printf(`I would've sent an email, but I won't...
FROM: %s
TO: %s
SUBJECT: %s

BODY:
%s
`, from, to, subject, body)
	} else {
		err = sendEmail(srv, from, to, subject, body)
		if err != nil {
			log.Fatalf("Failed to send email: %v", err)
		}
	}
}
