package meal_collection

import (
	"fmt"
	"log"
	"meals/calendar"
	"testing"
	"time"
)

const MEALS_JSON = "../data/recipes.json"

func TestMealCollectionReading(t *testing.T) {
	mealData, err := OpenMealData(MEALS_JSON)
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	_, err = ReadMealCollection(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}
}

func TestMealCollectionBadReading(t *testing.T) {
	mealData, err := OpenMealData("../data/recipes_with_unknown_field.json")
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	_, err = ReadMealCollection(mealData)
	expectedErr := "error unmarshalling JSON: json: unknown field \"unknown_field\""
	if err == nil || err.Error() != expectedErr {
		t.Errorf("Expected error: '%s', got: '%v'", expectedErr, err)
	}
}

func TestMealListGenerationFromCollection(t *testing.T) {
	mealData, err := OpenMealData(MEALS_JSON)
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	collection, err := ReadMealCollection(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	items := collection.GenerateMealsList(*calendar.NewCalendar(2024, time.October))
	if len(items) != 41 {
		t.Errorf("Expected length of list: '%d', got: '%d'", 41, len(items))
	}
}

func TestMealListGenerationUniquePerMonth(t *testing.T) {
	mealData, err := OpenMealData(MEALS_JSON)
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	collection, err := ReadMealCollection(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	itemsOctober := collection.GenerateMealsList(*calendar.NewCalendar(2024, time.October))
	itemsNovember := collection.GenerateMealsList(*calendar.NewCalendar(2024, time.November))

	// Items should match length within 2 items...
	if len(itemsOctober)-len(itemsNovember) > 2 || len(itemsOctober)-len(itemsNovember) < -2 {
		t.Errorf("Expected matching length: '%d', got: '%d'", len(itemsOctober), len(itemsNovember))
	}

	// But not match in ordering.
	allMatch := true
	for i := 0; i < len(itemsNovember); i++ {
		if itemsOctober[i].Name != itemsNovember[i].Name {
			allMatch = false
		}
	}
	if allMatch {
		t.Errorf("Both lists match, when they should not.")
	}
}

func TestMealListGenerationMatchAcrossMonth(t *testing.T) {
	mealData, err := OpenMealData(MEALS_JSON)
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	collection, err := ReadMealCollection(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	itemsOctober1 := collection.GenerateMealsList(*calendar.NewCalendar(2024, time.October))
	itemsOctober2 := collection.GenerateMealsList(*calendar.NewCalendar(2024, time.October))

	// Items should match length...
	if len(itemsOctober1) != len(itemsOctober2) {
		t.Errorf("Expected matching length: '%d', got: '%d'", len(itemsOctober1), len(itemsOctober2))
	}

	// And match on ordering.
	allMatch := true
	for i := 0; i < len(itemsOctober1); i++ {
		if itemsOctober1[i].Name != itemsOctober2[i].Name {
			fmt.Printf("These items do not match: '%s' '%s'", itemsOctober1[i].Name, itemsOctober2[i].Name)
			allMatch = false
		}
	}
	if !allMatch {
		t.Errorf("Both lists don't match, when they should.")
	}
}

func TestGenerateMealsWholeYearMatchAcrossMonth(t *testing.T) {
	mealData, err := OpenMealData(MEALS_JSON)
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	collection, err := ReadMealCollection(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	itemsOctober1 := collection.GenerateMealsWholeYear(*calendar.NewCalendar(2024, time.October))
	itemsOctober2 := collection.GenerateMealsWholeYear(*calendar.NewCalendar(2024, time.October))

	// Items should match length...
	if len(itemsOctober1) != len(itemsOctober2) {
		t.Errorf("Expected matching length: '%d', got: '%d'", len(itemsOctober1), len(itemsOctober2))
	}

	// And match on ordering.
	allMatch := true
	for i := 0; i < len(itemsOctober1); i++ {
		if itemsOctober1[i].Name != itemsOctober2[i].Name {
			fmt.Printf("These items do not match: '%s' '%s'", itemsOctober1[i].Name, itemsOctober2[i].Name)
			allMatch = false
		}
	}
	if !allMatch {
		t.Errorf("Both lists don't match, when they should.")
	}
}

func TestGenerateMealsWholeYearUniquePerMonth(t *testing.T) {
	mealData, err := OpenMealData(MEALS_JSON)
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	collection, err := ReadMealCollection(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	itemsOctober := collection.GenerateMealsWholeYear(*calendar.NewCalendar(2024, time.October))
	itemsNovember := collection.GenerateMealsWholeYear(*calendar.NewCalendar(2024, time.November))

	// Items should match length within 2 items...
	if len(itemsOctober)-len(itemsNovember) > 2 || len(itemsOctober)-len(itemsNovember) < -2 {
		t.Errorf("Expected matching length: '%d', got: '%d'", len(itemsOctober), len(itemsNovember))
	}

	// But not match in ordering.
	allMatch := true
	for i := 0; i < len(itemsNovember); i++ {
		if itemsOctober[i].Name != itemsNovember[i].Name {
			allMatch = false
		}
	}
	if allMatch {
		t.Errorf("Both lists match, when they should not.")
	}
}
