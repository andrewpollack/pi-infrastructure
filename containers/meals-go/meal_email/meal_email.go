package meal_email

import (
	"fmt"
	"meals/calendar"
	"meals/config"
	"meals/meal_collection"
	"strings"
	"time"
)

type EmailService int

const (
	Gmail = iota
	SES
)

type Date struct {
	Year  int
	Month int
	Day   int
}

type Config struct {
	PostgresURL    string
	EmailService   EmailService
	Sender         string
	Receivers      []string
	HardcodedMeals []string
	ExtraItems     []string
}

func (d Date) ToTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

func FromTime(t time.Time) Date {
	return Date{
		Year:  t.Year(),
		Month: int(t.Month()),
		Day:   t.Day(),
	}
}

func GetDaysOfCurrentWeek(date Date) []Date {
	t := date.ToTime()
	offset := int(t.Weekday())
	startOfWeek := t.AddDate(0, 0, -offset)

	var fullWeek []Date
	for i := 0; i < 7; i++ {
		day := startOfWeek.AddDate(0, 0, i)
		fullWeek = append(fullWeek, FromTime(day))
	}
	return fullWeek
}

func GetDaysOfNextWeek(date Date) []Date {
	t := date.ToTime()
	nextWeekStart := t.AddDate(0, 0, 7)
	return GetDaysOfCurrentWeek(FromTime(nextWeekStart))
}

type YearMonth struct {
	Year  int
	Month int
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
		currMeal := meals[i]

		if currMeal.URL != nil && *currMeal.URL != "" {
			sb.WriteString(fmt.Sprintf("            <td><a href='%s'>%s</a></td>\n", *currMeal.URL, currMeal.Name))
			continue
		} else {
			sb.WriteString(fmt.Sprintf("            <td>%s</td>\n", currMeal.Name))
		}
	}

	sb.WriteString(`        </tr>
    </tbody>
</table>

`)
	return sb.String()
}

func GenerateGroceryList(ingredients []meal_collection.Ingredient) string {
	var sb strings.Builder

	for _, aisle := range config.Cfg.App.Aisles {
		// Write a header for the aisle
		fmt.Fprintf(&sb, "<h4>%s</h4>\n", aisle)

		// Collect all items for this aisle
		var itemsForAisle []meal_collection.Ingredient
		for _, ing := range ingredients {
			if ing.Aisle == meal_collection.Aisle(aisle) {
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

func (c Config) GenerateEmailContentHTML(date Date, collection meal_collection.MealCollection, meals []meal_collection.Meal, ingredients []meal_collection.Ingredient) (string, error) {
	var sb strings.Builder
	sb.WriteString(generateHeader())
	sb.WriteString(generateTable(meals))
	sb.WriteString(GenerateGroceryList(ingredients))
	sb.WriteString(generateCloser())

	return sb.String(), nil
}

func (c Config) GetIngredientsForNextWeek(date Date, collection meal_collection.MealCollection) ([]meal_collection.Ingredient, error) {
	var ingredients []meal_collection.Ingredient

	allMeals, err := c.GetMealsForNextWeek(date, collection)
	if err != nil {
		return ingredients, fmt.Errorf("failed to get meals for next week: %v", err)
	}
	allExtraItems, err := c.GetExtraItems()
	if err != nil {
		return ingredients, fmt.Errorf("failed to get extra items: %v", err)
	}

	ingredients = meal_collection.MealsToIngredients(allMeals)
	for _, extraItem := range allExtraItems {
		ingredients = append(ingredients, meal_collection.ExtraItemToIngredient(extraItem))
	}

	return ingredients, nil
}

func (c Config) GetMealsForNextWeek(date Date, collection meal_collection.MealCollection) ([]meal_collection.Meal, error) {
	var allMeals []meal_collection.Meal

	// Decide how to get meals: either hardcoded or generated
	if len(c.HardcodedMeals) == 7 {
		fullCollection, err := meal_collection.ReadMealCollectionFromDB(c.PostgresURL, time.Now().Unix())
		if err != nil {
			return nil, fmt.Errorf("something went wrong reading meals: %s", err)
		}
		mealMap := fullCollection.MapNameToMeal()
		for _, v := range c.HardcodedMeals {
			meal, ok := mealMap[v]
			if !ok {
				return nil, fmt.Errorf("meal not found: %s", v)
			}

			allMeals = append(allMeals, meal)
		}
	} else {
		daysOfWeek := GetDaysOfNextWeek(date)
		calendars := make(map[YearMonth][]meal_collection.Meal)

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

	return allMeals, nil
}

func (c Config) GetExtraItems() ([]meal_collection.ExtraItem, error) {
	if len(c.ExtraItems) == 0 {
		return nil, nil
	}
	extraItemsDB, err := meal_collection.ReadExtraItemsFromDB(c.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to read extra items from DB: %v", err)
	}

	extraItemsMap := map[string]meal_collection.ExtraItem{}
	for _, extraItem := range extraItemsDB {
		extraItemsMap[extraItem.Name] = extraItem
	}
	var extraItems []meal_collection.ExtraItem
	for _, extraItem := range c.ExtraItems {
		// If the extraItem isn't found in the map, return an error.
		if _, found := extraItemsMap[extraItem]; !found {
			return nil, fmt.Errorf("extra item not found: %s", extraItem)
		}

		extraItems = append(extraItems, extraItemsMap[extraItem])
	}

	return extraItems, nil
}

func GenerateHeaderForNextWeek(date Date) string {
	daysOfWeek := GetDaysOfNextWeek(date)
	first := daysOfWeek[0]
	last := daysOfWeek[6]

	return fmt.Sprintf("Meals for %s %d -> %s %d ", time.Month(first.Month), first.Day, time.Month(last.Month), last.Day)
}

func (c Config) CreateAndSendEmail() error {
	now := time.Now()

	// 1) Read the meal collection
	collection, err := meal_collection.ReadMealCollectionFromDB(c.PostgresURL, now.Unix())
	if err != nil {
		return fmt.Errorf("failed to read meals from DB: %w", err)
	}

	// 2) Get the next week’s meals/ingredients
	currDate := Date{
		Year:  now.Year(),
		Month: int(now.Month()),
		Day:   now.Day(),
	}
	meals, err := c.GetMealsForNextWeek(currDate, collection)
	if err != nil {
		return fmt.Errorf("failed to get meals for next week: %w", err)
	}

	ingredients, err := c.GetIngredientsForNextWeek(currDate, collection)
	if err != nil {
		return fmt.Errorf("failed to get ingredients for next week: %w", err)
	}

	// 3) Build email subject and HTML body
	subject := GenerateHeaderForNextWeek(currDate)
	bodyHTML, err := c.GenerateEmailContentHTML(currDate, collection, meals, ingredients)
	if err != nil {
		return fmt.Errorf("failed to generate email HTML: %w", err)
	}

	// 4) Generate PDF attachment
	pdfBytes, err := DefaultPDFGenerator{}.GenerateIngredientsPDF(ingredients)
	if err != nil {
		return fmt.Errorf("failed to generate ingredients PDF: %w", err)
	}

	// 5) Generating the PDF name as the first day of the next week
	nextWeekDays := GetDaysOfNextWeek(currDate)
	if len(nextWeekDays) == 0 {
		return fmt.Errorf("GetDaysOfNextWeek returned no days")
	}
	first := nextWeekDays[0]
	pdfName := fmt.Sprintf("%d-%02d-%02d-grocery-list.pdf", first.Year, first.Month, first.Day)

	var sender EmailSender
	switch c.EmailService {
	case SES:
		sender = SESEmailSender{
			From: c.Sender,
			To:   c.Receivers,
		}
	case Gmail:
		gs, err := AuthenticateGmail()
		if err != nil {
			return fmt.Errorf("failed to authenticate with Gmail: %w", err)
		}
		sender = GmailSender{
			From:    c.Sender,
			To:      c.Receivers,
			Service: gs,
		}
	default:
		return fmt.Errorf("unsupported email service: %d", c.EmailService)
	}

	err = sender.SendEmail(subject, bodyHTML, pdfBytes, pdfName)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
