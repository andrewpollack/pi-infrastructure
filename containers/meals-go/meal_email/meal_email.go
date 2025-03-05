package meal_email

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"meals/calendar"
	"meals/meal_collection"
	"mime/quotedprintable"
	"net/textproto"
	"os"
	"strings"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
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

func GenerateGroceryList(meals []meal_collection.Meal) string {
	var sb strings.Builder

	ingredients := meal_collection.MealsToIngredients(meals)

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
	allMeals := GetMealsForNextWeek(date, collection)

	var sb strings.Builder
	sb.WriteString(generateHeader())
	sb.WriteString(generateTable(allMeals))
	sb.WriteString(GenerateGroceryList(allMeals))
	sb.WriteString(generateCloser())

	return sb.String()
}

func GetMealsForNextWeek(date Date, collection meal_collection.MealCollection) []meal_collection.Meal {
	daysOfWeek := GetDaysOfNextWeek(date)
	calendars := make(map[YearMonth][]meal_collection.Meal)
	var allMeals []meal_collection.Meal

	// Decide how to get meals: either hardcoded or generated
	_, useHardcoded := os.LookupEnv("H_1")
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

	return allMeals
}

func GenerateHeaderForNextWeek(date Date) string {
	daysOfWeek := GetDaysOfNextWeek(date)
	first := daysOfWeek[0]
	last := daysOfWeek[6]

	return fmt.Sprintf("Meals for %s %d -> %s %d ", time.Month(first.Month), first.Day, time.Month(last.Month), last.Day)
}

func convertHTMLToPDF(html string) ([]byte, error) {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF generator: %v", err)
	}

	pdfg.MarginLeft.Set(0)
	pdfg.MarginRight.Set(0)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.Cover.Zoom.Set(1.0)

	page := wkhtmltopdf.NewPageReader(strings.NewReader(html))
	pdfg.AddPage(page)

	if err := pdfg.Create(); err != nil {
		return nil, fmt.Errorf("failed to create PDF: %v", err)
	}

	return pdfg.Bytes(), nil
}

func generateIngredientsPDF(meals []meal_collection.Meal) ([]byte, error) {
	ingredients := meal_collection.MealsToIngredients(meals)

	htmlContent := buildHTMLContent(ingredients)
	pdfBytes, err := convertHTMLToPDF(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("error converting HTML to PDF: %w", err)
	}
	return pdfBytes, nil
}

func buildHTMLContent(ingredients []meal_collection.Ingredient) string {
	const CELLS_PER_ROW = 3

	// Preallocate an estimated capacity for the builder.
	var sb strings.Builder
	sb.Grow(1024)

	// Write the header of the HTML document.
	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Aisle Items</title>
	<style>
		html, body {
			margin: 0;
			padding: 0;
			width: 100%;
			height: 100%;
		}
		table {
			width: 100%;
			height: 100%;
			border-collapse: collapse;
			table-layout: fixed; /* Forces fixed column widths */
		}
		tr {
			height: calc(100% / 3); /* Each row is 1/3 of the page height */
		}
		td {
			width: 33.33%; /* Each column takes 1/3 of the table width */
			border: 1px solid #333;
			background-color: #ffffff;
			vertical-align: top;
			padding: 5px;
		}
		h3 {
			margin: 0 0 5px;
			padding: 2px;
			background-color: #00d5ff;
			text-align: center;
			font-size: 14px;
		}
		.checkbox-group label {
			font-size: 12px;
			margin: 0; /* Remove extra margin */
			padding: 0; /* Remove extra padding */
		}
		input[type="checkbox"] {
			width: 12px;
			height: 12px;
			margin: 0; /* Remove default checkbox margin */
			padding: 0;
		}
	</style>
</head>
<body>
<table>
`)

	// Generate table rows: three aisles per row.
	for i, aisle := range meal_collection.AllAisles {
		if i%CELLS_PER_ROW == 0 {
			sb.WriteString("  <tr>\n")
		}

		aisleHTML := buildAisleCellHTML(aisle, ingredients)
		sb.WriteString(aisleHTML)

		if i%CELLS_PER_ROW == CELLS_PER_ROW-1 {
			sb.WriteString("  </tr>\n")
		}
	}

	// Write closing tags.
	sb.WriteString(`</table>
</body>
</html>`)

	return sb.String()
}

func buildAisleCellHTML(aisle meal_collection.Aisle, ingredients []meal_collection.Ingredient) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("    <td>\n      <h3>%s</h3>\n", aisle))

	// Filter ingredients for the current aisle.
	var itemsForAisle []meal_collection.Ingredient
	for _, ing := range ingredients {
		if ing.Aisle == aisle {
			itemsForAisle = append(itemsForAisle, ing)
		}
	}

	if len(itemsForAisle) > 0 {
		sb.WriteString("      <div class=\"checkbox-group\">\n")
		for _, ing := range itemsForAisle {
			sb.WriteString(fmt.Sprintf("        <label><input type=\"checkbox\" disabled> %s</label><br>\n", ing.StringBolded()))
		}
		sb.WriteString("      </div>\n")
	}

	sb.WriteString("    </td>\n")
	return sb.String()
}

func CreateAndSendEmail(useSES bool) error {
	currentTime := time.Now()

	// Extract the year, month, and day
	year := currentTime.Year()
	month := currentTime.Month()
	day := currentTime.Day()
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, currentTime.Location())

	collection, err := meal_collection.ReadMealCollectionFromDB(firstOfMonth.Unix())
	if err != nil {
		return fmt.Errorf("something went wrong reading meals: %s", err)
	}

	currDate := Date{Year: year, Month: int(month), Day: day}
	from := os.Getenv("SENDER_EMAIL")
	to := os.Getenv("RECEIVER_EMAILS")
	subject := GenerateHeaderForNextWeek(currDate)
	bodyHtml := GenerateEmailForNextWeek(currDate, collection)

	dryRun := os.Getenv("DRY_RUN")
	if dryRun == "true" {
		fmt.Printf(`I would've sent an email, but I won't...
FROM: %s
TO: %s
SUBJECT: %s

BODY:
%s
`, from, to, subject, bodyHtml)
		return nil
	}

	if useSES {
		meals := GetMealsForNextWeek(currDate, collection)
		pdfBytes, err := generateIngredientsPDF(meals)
		if err != nil {
			return fmt.Errorf("failed to generate ingredients PDF: %v", err)
		}

		err = sendEmailSESWithAttachmentBytes(from, to, subject, bodyHtml, pdfBytes, "grocery_list.pdf")
		if err != nil {
			return fmt.Errorf("failed to send SES email: %v", err)
		}
	} else {
		gs, err := AuthenticateGmail()
		if err != nil {
			return fmt.Errorf("failed to authenticate with Gmail: %s", err.Error())
		}

		err = gs.SendEmail(from, to, subject, bodyHtml)
		if err != nil {
			return fmt.Errorf("failed to send Gmail email: %v", err)
		}
	}

	return nil
}

func sendEmailSESWithAttachmentBytes(from, to, subject, bodyHtml string, attachmentBytes []byte, attachmentFilename string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}
	client := ses.NewFromConfig(cfg)

	// Process multiple recipients.
	recipientList := []string{}
	for _, addr := range strings.Split(to, ",") {
		trimmed := strings.TrimSpace(addr)
		if trimmed != "" {
			recipientList = append(recipientList, trimmed)
		}
	}
	toHeader := strings.Join(recipientList, ", ")

	var emailRaw bytes.Buffer
	boundaryMixed := "NextPartMixedBoundary"
	boundaryAlternative := "NextPartAlternativeBoundary"

	// Headers
	emailRaw.WriteString(fmt.Sprintf("From: %s\r\n", from))
	emailRaw.WriteString(fmt.Sprintf("To: %s\r\n", toHeader))
	emailRaw.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	emailRaw.WriteString("MIME-Version: 1.0\r\n")
	emailRaw.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundaryMixed))
	emailRaw.WriteString("\r\n") // End headers

	// Start multipart/mixed section.
	emailRaw.WriteString(fmt.Sprintf("--%s\r\n", boundaryMixed))
	// Create multipart/alternative section for HTML content.
	emailRaw.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundaryAlternative))
	emailRaw.WriteString("\r\n")

	// Add the HTML part.
	emailRaw.WriteString(fmt.Sprintf("--%s\r\n", boundaryAlternative))
	emailRaw.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	emailRaw.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	emailRaw.WriteString("\r\n")
	qp := quotedprintable.NewWriter(&emailRaw)
	_, err = qp.Write([]byte(bodyHtml))
	if err != nil {
		return fmt.Errorf("failed to write html body: %v", err)
	}
	qp.Close()
	emailRaw.WriteString("\r\n")
	// End alternative part.
	emailRaw.WriteString(fmt.Sprintf("--%s--\r\n", boundaryAlternative))

	// Add attachment part.
	emailRaw.WriteString(fmt.Sprintf("--%s\r\n", boundaryMixed))
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", "application/octet-stream")
	h.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachmentFilename))
	h.Set("Content-Transfer-Encoding", "base64")
	for key, vals := range h {
		for _, v := range vals {
			emailRaw.WriteString(fmt.Sprintf("%s: %s\r\n", key, v))
		}
	}
	emailRaw.WriteString("\r\n")
	encodedAttachment := base64.StdEncoding.EncodeToString(attachmentBytes)
	// RFC 2045 recommends splitting base64 into 76-character lines.
	for i := 0; i < len(encodedAttachment); i += 76 {
		end := i + 76
		if end > len(encodedAttachment) {
			end = len(encodedAttachment)
		}
		emailRaw.WriteString(encodedAttachment[i:end] + "\r\n")
	}
	emailRaw.WriteString(fmt.Sprintf("--%s--\r\n", boundaryMixed))

	// Prepare the raw email message.
	rawMessage := types.RawMessage{
		Data: emailRaw.Bytes(),
	}

	input := &ses.SendRawEmailInput{
		RawMessage: &rawMessage,
	}

	_, err = client.SendRawEmail(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to send raw email: %v", err)
	}

	return nil
}
