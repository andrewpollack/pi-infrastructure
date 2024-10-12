package meal_email

import (
	"fmt"
	"log"
	"math"
	"meals/calendar"
	"meals/meal_collection"
	"os"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

type Aisle string

// Define constants for all the possible values of Aisle
const (
	AisleCheeseAndBakery     string = "Cheese & Bakery"
	AisleAlcoholButterCheese string = "18 & 19 (Alcohol, Butter, Cheese)"
	AisleFreezer             string = "16 & 17 (Freezer)"
	AisleNoFoodItems         string = "10-15 (No Food Items)"
	AisleBeveragesAndSnacks  string = "6-9 (Bevs & Snacks)"
	AisleBreakfastAndBaking  string = "3-5 (Breakfast & Baking)"
	AislePastaGlobalCanned   string = "1 & 2 (Pasta, Global, Canned)"
	AisleProduce             string = "Produce"
	AisleMeatAndYogurt       string = "Meat & Yogurt"
)

// AllAisles contains the list of all valid aisle values
var AllAisles = []string{
	AisleCheeseAndBakery,
	AisleAlcoholButterCheese,
	AisleFreezer,
	AisleNoFoodItems,
	AisleBeveragesAndSnacks,
	AisleBreakfastAndBaking,
	AislePastaGlobalCanned,
	AisleProduce,
	AisleMeatAndYogurt,
}

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

func GenerateGroceryCollection(meals []meal_collection.Item) map[string]map[string]meal_collection.GroceryItem {
	// Aisle -> ITEM__UNIT -> QUANTITY, RECIPES
	groceryCollection := make(map[string]map[string]meal_collection.GroceryItem)

	for _, aisle := range AllAisles {
		groceryCollection[aisle] = make(map[string]meal_collection.GroceryItem)
	}

	for _, currItem := range meals {
		if len(currItem.Ingredients) == 0 {
			continue
		}

		for _, ingredient := range currItem.Ingredients {
			aisle := string(ingredient.Aisle)
			item := ingredient.Item
			quantity := ingredient.Quantity
			unit := ingredient.Unit

			// Key combining item and unit
			k := fmt.Sprintf("%s__%s", item, unit)

			if _, exists := groceryCollection[aisle][k]; !exists {
				groceryCollection[aisle][k] = meal_collection.GroceryItem{Quantity: 0, RelatedMeals: []string{}}
			}

			quant := groceryCollection[aisle][k].Quantity + quantity
			relatedMeals := append(groceryCollection[aisle][k].RelatedMeals, currItem.Name)

			groceryCollection[aisle][k] = meal_collection.GroceryItem{Quantity: quant, RelatedMeals: relatedMeals}
		}
	}

	return groceryCollection
}

func formatSignificant(f float64, sigDigits int) string {
	if f == 0 {
		return "0"
	}

	// Calculate the order of magnitude (log10) to scale the float
	scale := math.Pow10(sigDigits - 1 - int(math.Floor(math.Log10(math.Abs(f)))))

	// Scale, round, and rescale the float
	result := math.Round(f*scale) / scale

	// Format the result with %g to drop unnecessary trailing zeroes
	return fmt.Sprintf("%g", result)
}

func CreateGroceryEmailMessage(meals []meal_collection.Item) string {
	groceryCollection := GenerateGroceryCollection(meals)

	groceryEmail := ""
	for _, aisle := range AllAisles {
		items := []string{}

		itemsForIsle, ok := groceryCollection[aisle]
		if !ok {
			continue
		}

		// Get the keys (item names) and sort them
		keys := make([]string, 0, len(itemsForIsle))
		for k := range itemsForIsle {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			// Aisle -> ITEM__UNIT -> QUANTITY, RECIPES
			v := itemsForIsle[k]
			quant := v.Quantity
			relatedMeals := strings.Join(v.RelatedMeals, ", ")
			parts := strings.Split(k, "__")
			item := parts[0]
			unit := parts[1]

			// Format the item as * quantity unit: item (related_meals)
			items = append(items, fmt.Sprintf("%s %s: %s (%s)", formatSignificant(quant, 4), unit, item, relatedMeals))
		}

		groceryEmail += fmt.Sprintf("<h4>%s</h4>\n", aisle)

		if len(items) == 0 {
			// If there are no items, display "NONE"
			groceryEmail += "<p>NONE</p>\n"
		} else {
			// Start an unordered list for the items
			groceryEmail += "<ul style='margin-left: 20px;'>\n"
			for _, item := range items {
				groceryEmail += fmt.Sprintf("<li>%s</li>\n", item)
			}
			groceryEmail += "</ul>\n"
		}

		groceryEmail += "<br>\n"
	}

	return groceryEmail
}

func GenerateEmailForNextWeek(date Date, collection meal_collection.MealCollection) string {
	daysOfWeek := GetDaysOfNextWeek(date)

	calendars := make(map[YearMonth][]meal_collection.Item)

	var allItems []meal_collection.Item
	for _, day := range daysOfWeek {
		currYearMonth := YearMonth{day.Year, day.Month}

		if _, exists := calendars[currYearMonth]; !exists {
			calendars[currYearMonth] = collection.GenerateMealsList(*calendar.NewCalendar(day.Year, time.Month(day.Month)))
		}

		allItems = append(allItems, calendars[currYearMonth][day.Day-1])
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

	return fmt.Sprintf("Meals for %s %d → %s %d ", time.Month(first.Month), first.Day, time.Month(last.Month), last.Day)
}

func CreateAndSendEmail(srv *gmail.Service) {
	currentTime := time.Now()

	// Extract the year, month, and day
	year := currentTime.Year()
	month := currentTime.Month()
	day := currentTime.Day()

	collection, err := meal_collection.ReadMealCollection("")
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
