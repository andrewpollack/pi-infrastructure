package meal_email

import (
	"log"
	"meals/calendar"
	"meals/meal_collection"
	"testing"
	"time"
)

const MEALS_JSON = "../data/recipes.json"

func TestWeekGeneration(t *testing.T) {
	type DayToExpectedIndex struct {
		Day          Date
		ExpectedWeek []Date
	}

	tuples := []DayToExpectedIndex{
		{
			Day: Date{
				Year:  2024,
				Month: 10,
				Day:   7,
			},
			ExpectedWeek: []Date{
				{Year: 2024, Month: 10, Day: 6},
				{Year: 2024, Month: 10, Day: 7},
				{Year: 2024, Month: 10, Day: 8},
				{Year: 2024, Month: 10, Day: 9},
				{Year: 2024, Month: 10, Day: 10},
				{Year: 2024, Month: 10, Day: 11},
				{Year: 2024, Month: 10, Day: 12},
			},
		},
		{
			Day: Date{
				Year:  2024,
				Month: 10,
				Day:   28,
			},
			ExpectedWeek: []Date{
				{Year: 2024, Month: 10, Day: 27},
				{Year: 2024, Month: 10, Day: 28},
				{Year: 2024, Month: 10, Day: 29},
				{Year: 2024, Month: 10, Day: 30},
				{Year: 2024, Month: 10, Day: 31},
				{Year: 2024, Month: 11, Day: 1},
				{Year: 2024, Month: 11, Day: 2},
			},
		},
	}

	for _, tuple := range tuples {
		// Get the week index of the given day
		weekCollection := GetDaysOfCurrentWeek(tuple.Day)

		// Assert that the returned week matches the expected value
		if len(weekCollection) != len(tuple.ExpectedWeek) {
			t.Errorf("For day %d, expected %d dates, but got %d", tuple.Day, len(tuple.ExpectedWeek), len(weekCollection))
			continue
		}

		// Compare each day in the week
		for i := range weekCollection {
			if weekCollection[i] != tuple.ExpectedWeek[i] {
				t.Errorf("For day %d, expected date %v, but got %v", tuple.Day.Day, tuple.ExpectedWeek[i], weekCollection[i])
			}
		}
	}
}

func TestNextWeekGeneration(t *testing.T) {
	type DayToExpectedIndex struct {
		Day          Date
		ExpectedWeek []Date
	}

	tuples := []DayToExpectedIndex{
		{
			Day: Date{
				Year:  2024,
				Month: 10,
				Day:   7,
			},
			ExpectedWeek: []Date{
				{Year: 2024, Month: 10, Day: 13},
				{Year: 2024, Month: 10, Day: 14},
				{Year: 2024, Month: 10, Day: 15},
				{Year: 2024, Month: 10, Day: 16},
				{Year: 2024, Month: 10, Day: 17},
				{Year: 2024, Month: 10, Day: 18},
				{Year: 2024, Month: 10, Day: 19},
			},
		},
		{
			Day: Date{
				Year:  2024,
				Month: 10,
				Day:   28,
			},
			ExpectedWeek: []Date{
				{Year: 2024, Month: 11, Day: 3},
				{Year: 2024, Month: 11, Day: 4},
				{Year: 2024, Month: 11, Day: 5},
				{Year: 2024, Month: 11, Day: 6},
				{Year: 2024, Month: 11, Day: 7},
				{Year: 2024, Month: 11, Day: 8},
				{Year: 2024, Month: 11, Day: 9},
			},
		},
		{
			Day: Date{
				Year:  2024,
				Month: 11,
				Day:   25,
			},
			ExpectedWeek: []Date{
				{Year: 2024, Month: 12, Day: 1},
				{Year: 2024, Month: 12, Day: 2},
				{Year: 2024, Month: 12, Day: 3},
				{Year: 2024, Month: 12, Day: 4},
				{Year: 2024, Month: 12, Day: 5},
				{Year: 2024, Month: 12, Day: 6},
				{Year: 2024, Month: 12, Day: 7},
			},
		},
		{
			Day: Date{
				Year:  2024,
				Month: 12,
				Day:   30,
			},
			ExpectedWeek: []Date{
				{Year: 2025, Month: 1, Day: 5},
				{Year: 2025, Month: 1, Day: 6},
				{Year: 2025, Month: 1, Day: 7},
				{Year: 2025, Month: 1, Day: 8},
				{Year: 2025, Month: 1, Day: 9},
				{Year: 2025, Month: 1, Day: 10},
				{Year: 2025, Month: 1, Day: 11},
			},
		},
	}

	for _, tuple := range tuples {
		// Get the week index of the given day
		weekCollection := GetDaysOfNextWeek(tuple.Day)

		// Assert that the returned week matches the expected value
		if len(weekCollection) != len(tuple.ExpectedWeek) {
			t.Errorf("For day %d, expected %d dates, but got %d", tuple.Day, len(tuple.ExpectedWeek), len(weekCollection))
			continue
		}

		// Compare each day in the week
		for i := range weekCollection {
			if weekCollection[i] != tuple.ExpectedWeek[i] {
				t.Errorf("For day %d, expected date %v, but got %v", tuple.Day.Day, tuple.ExpectedWeek[i], weekCollection[i])
			}
		}
	}
}

func TestGroceryListGeneration(t *testing.T) {
	mealData, err := meal_collection.OpenMealData(MEALS_JSON)
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	collection, err := meal_collection.ReadMealCollectionFromReader(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	itemsOctober := collection.GenerateMealsList(*calendar.NewCalendar(2024, time.October))
	_ = GenerateGroceryList(itemsOctober)

	// TODO: Actually test here. Golden tests are a pain comparing against a changing output...
}

func TestEmailGeneration(t *testing.T) {
	mealData, err := meal_collection.OpenMealData(MEALS_JSON)
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	collection, err := meal_collection.ReadMealCollectionFromReader(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	_ = GenerateEmailForNextWeek(Date{Year: 2024, Month: 10, Day: 28}, collection)
	// TODO: Actually test here. Golden tests are a pain comparing against a changing output...
}
