package meal_calendar

import (
	"meals/calendar"
	"meals/meal_collection"
	"testing"
	"time"
)

func TestCalendarHTMLGeneration(t *testing.T) {
	collection, err := meal_collection.ReadMealCollection("../data/recipes.json")
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	curr_meal_calender := NewCalendar(*calendar.NewCalendar(2024, time.February), collection)
	_ = curr_meal_calender.RenderHTMLCalendar()

	// TODO: Actually test here. Golden tests are a pain comparing against a changing
	// output...
}
