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

	var allMeals []meal_collection.Meal
	for i, v := range arr {
		if i == 4 {
			allMeals = append(allMeals, meal_collection.MEAL_LEFTOVERS)
			continue
		}
		if i == 5 {
			allMeals = append(allMeals, meal_collection.MEAL_OUT)
			continue
		}

		for _, fullItem := range flattenedItems {
			if fullItem.Name == v {
				allMeals = append(allMeals, fullItem)
				break
			}
		}
	}

	return allMeals
}

func generateHeader() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>Meals for Next Week</title>
</head>
<body>
	<h3>Meals:</h3>
`
}

func generateTable(meals []meal_collection.Meal) string {
	var sb strings.Builder
	sb.WriteString(`<table border='1'>
<thead>
<tr>`)

	fullDaysOfWeek := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	for _, day := range fullDaysOfWeek {
		sb.WriteString(fmt.Sprintf("            <th>%s</th>\n", day))
	}

	sb.WriteString(`        </tr>
    </thead>
    <tbody>
        <tr>
`)

	for i := range fullDaysOfWeek {
		sb.WriteString(fmt.Sprintf("            <td>%s</td>\n", meals[i].Name))
	}

	sb.WriteString(`        </tr>
    </tbody>
</table>

`)
	return sb.String()
}

func GenerateGroceryList(meals []meal_collection.Meal) string {
	var sb strings.Builder

	ingredients := meal_collection.MealsToIngredients(meals)
	sort.Slice(ingredients, func(i, j int) bool {
		return ingredients[i].Name < ingredients[j].Name
	})

	for _, aisle := range meal_collection.AllAisles {
		// Write a header for the aisle
		fmt.Fprintf(&sb, "<h4>%s</h4>\n", aisle)

		// Collect all items for this aisle
		var itemsForAisle []meal_collection.Ingredient
		for _, ing := range ingredients {
			if ing.Aisle == aisle {
				itemsForAisle = append(itemsForAisle, ing)
			}
		}

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

func generateCloser() string {
	return `
	</body>
</html>`
}

func GenerateEmailForNextWeek(date Date, collection meal_collection.MealCollection) string {
	daysOfWeek := GetDaysOfNextWeek(date)
	calendars := make(map[YearMonth][]meal_collection.Meal)
	var allMeals []meal_collection.Meal

	// Decide how to get meals: either hardcoded or generated
	useHardcoded := os.Getenv("USE_HARDCODE") == "true"
	if useHardcoded {
		allMeals = useHardcodedValues(collection)
	} else {
		for _, day := range daysOfWeek {
			currYearMonth := YearMonth{Year: day.Year, Month: day.Month}
			if _, exists := calendars[currYearMonth]; !exists {
				// Generate all meals for the entire year/month if not already present
				c := calendar.NewCalendar(day.Year, time.Month(day.Month))
				calendars[currYearMonth] = collection.GenerateMealsWholeYearNoCategories(*c)
			}
			allMeals = append(allMeals, calendars[currYearMonth][day.Day-1])
		}
	}

	var sb strings.Builder
	sb.WriteString(generateHeader())
	sb.WriteString(generateTable(allMeals))
	sb.WriteString(GenerateGroceryList(allMeals))
	sb.WriteString(generateCloser())

	return sb.String()
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

	collection, err := meal_collection.ReadMealCollectionFromDB()
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
