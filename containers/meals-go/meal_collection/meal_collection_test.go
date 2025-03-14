package meal_collection

import (
	"log"
	"meals/calendar"
	"reflect"
	"sort"
	"testing"
	"time"
)

const MEALS_JSON = "../data/recipes.json"

func TestMealCollectionReading(t *testing.T) {
	mealData, err := OpenMealData(MEALS_JSON)
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	_, err = ReadMealCollectionFromReader(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}
}

func TestMealCollectionBadReading(t *testing.T) {
	mealData, err := OpenMealData("../data/recipes_with_unknown_field.json")
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	_, err = ReadMealCollectionFromReader(mealData)
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

	collection, err := ReadMealCollectionFromReader(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	items := collection.GenerateMealsWholeYearNoCategories(*calendar.NewCalendar(2024, time.October))
	if len(items) != 31 {
		t.Errorf("Expected length of list: '%d', got: '%d'", 31, len(items))
	}
}

func TestMealListGenerationUniquePerMonth(t *testing.T) {
	mealData, err := OpenMealData(MEALS_JSON)
	if err != nil {
		log.Fatalf("Error fetching mealData: %v", err)
	}

	collection, err := ReadMealCollectionFromReader(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	itemsOctober := collection.GenerateMealsWholeYearNoCategories(*calendar.NewCalendar(2024, time.October))
	itemsNovember := collection.GenerateMealsWholeYearNoCategories(*calendar.NewCalendar(2024, time.November))

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

	collection, err := ReadMealCollectionFromReader(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	itemsOctober1 := collection.GenerateMealsWholeYearNoCategories(*calendar.NewCalendar(2024, time.October))
	itemsOctober2 := collection.GenerateMealsWholeYearNoCategories(*calendar.NewCalendar(2024, time.October))

	// Items should match length...
	if len(itemsOctober1) != len(itemsOctober2) {
		t.Errorf("Expected matching length: '%d', got: '%d'", len(itemsOctober1), len(itemsOctober2))
	}

	// And match on ordering.
	allMatch := true
	for i := 0; i < len(itemsOctober1); i++ {
		if itemsOctober1[i].Name != itemsOctober2[i].Name {
			log.Printf("These items do not match: '%s' '%s'", itemsOctober1[i].Name, itemsOctober2[i].Name)
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

	collection, err := ReadMealCollectionFromReader(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	itemsOctober1 := collection.GenerateMealsWholeYearNoCategories(*calendar.NewCalendar(2024, time.October))
	itemsOctober2 := collection.GenerateMealsWholeYearNoCategories(*calendar.NewCalendar(2024, time.October))

	// Items should match length...
	if len(itemsOctober1) != len(itemsOctober2) {
		t.Errorf("Expected matching length: '%d', got: '%d'", len(itemsOctober1), len(itemsOctober2))
	}

	// And match on ordering.
	allMatch := true
	for i := 0; i < len(itemsOctober1); i++ {
		if itemsOctober1[i].Name != itemsOctober2[i].Name {
			log.Printf("These items do not match: '%s' '%s'", itemsOctober1[i].Name, itemsOctober2[i].Name)
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

	collection, err := ReadMealCollectionFromReader(mealData)
	if err != nil {
		t.Errorf("Something went wrong reading meals... %s", err)
	}

	itemsOctober := collection.GenerateMealsWholeYearNoCategories(*calendar.NewCalendar(2024, time.October))
	itemsNovember := collection.GenerateMealsWholeYearNoCategories(*calendar.NewCalendar(2024, time.November))

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

func TestMealsToIngredients(t *testing.T) {
	// Sample input data
	meals := []Meal{
		{
			Name: "Pizza",
			Ingredients: []Ingredient{
				{Name: "Cheese", Quantity: 2, Unit: UnitCount, Aisle: AisleCheeseAndBakery},
				{Name: "Tomato Sauce", Quantity: 1, Unit: UnitCup, Aisle: AisleProduce},
			},
		},
		{
			Name: "Burger",
			Ingredients: []Ingredient{
				{Name: "Cheese", Quantity: 2, Unit: UnitCount, Aisle: AisleCheeseAndBakery},
				{Name: "Bun", Quantity: 4, Unit: UnitCount, Aisle: AisleCheeseAndBakery},
				{Name: "Beef", Quantity: 1, Unit: UnitLb, Aisle: AisleMeatAndYogurt},
			},
		},
	}

	got := MealsToIngredients(meals)
	want := []Ingredient{
		{
			Name:         "Bun",
			Unit:         UnitCount,
			Quantity:     4,
			Aisle:        AisleCheeseAndBakery,
			RelatedMeals: []string{"Burger"},
		},
		{
			Name:         "Cheese",
			Unit:         UnitCount,
			Quantity:     4, // 2 + 2
			Aisle:        AisleCheeseAndBakery,
			RelatedMeals: []string{"Burger", "Pizza"},
		},
		{
			Name:         "Tomato Sauce",
			Unit:         UnitCup,
			Quantity:     1,
			Aisle:        AisleProduce,
			RelatedMeals: []string{"Pizza"},
		},
		{
			Name:         "Beef",
			Unit:         UnitLb,
			Quantity:     1,
			Aisle:        AisleMeatAndYogurt,
			RelatedMeals: []string{"Burger"},
		},
	}

	sortIngredients(got)
	sortIngredients(want)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("MealsToIngredients() = %#v;\nwant %#v", got, want)
	}
}

// sortIngredients sorts by Aisle, then by Name, for consistent comparison.
func sortIngredients(ings []Ingredient) {
	sort.Slice(ings, func(i, j int) bool {
		if ings[i].Aisle != ings[j].Aisle {
			return ings[i].Aisle < ings[j].Aisle
		}
		return ings[i].Name < ings[j].Name
	})
}
