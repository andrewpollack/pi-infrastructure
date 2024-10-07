package meal_calendar

import (
	"fmt"
	"meals/calendar"
	"meals/meal_collection"
)

// MealCalendar combines calendar and meal collection functionalities
type MealCalendar struct {
	Calendar       calendar.Calendar
	MealCollection meal_collection.MealCollection
}

func NewCalendar(calendar calendar.Calendar, meal_collection meal_collection.MealCollection) *MealCalendar {
	meal_calendar := &MealCalendar{
		Calendar:       calendar,
		MealCollection: meal_collection,
	}

	return meal_calendar
}

// Example method that uses both Calendar and MealCollection
func (mc *MealCalendar) RenderHTMLCalendar() string {
	items := mc.MealCollection.GenerateMealsList(mc.Calendar)

	mealCalendar := fmt.Sprintf(`
<h1>%s %d</h1>
<table>
	<tr>
		<th>Sunday</th>
		<th>Monday</th>
		<th>Tuesday</th>
		<th>Wednesday</th>
		<th>Thursday</th>
		<th>Friday</th>
		<th>Saturday</th>
	</tr>
`, mc.Calendar.Month, mc.Calendar.Year)

	for _, week := range mc.Calendar.Weeks {
		mealCalendar += "\t<tr>\n"
		for _, day := range week {
			mealCalendar += "\t\t<td>"
			if day.Number == 0 {
				mealCalendar += "NONE"
			} else {
				item := items[day.Number-1]
				if item.URL != nil {
					mealCalendar += fmt.Sprintf("<a href=\"%s\">%s</a>", *item.URL, item.Name)
				} else {
					mealCalendar += item.Name
				}
			}
			mealCalendar += "</td>\n"
		}
		mealCalendar += "\t</tr>\n"
	}

	mealCalendar += `
</table>
`
	return mealCalendar
}
