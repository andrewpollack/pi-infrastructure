package meal_collection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"meals/calendar"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/exp/rand"
)

type Aisle string

// Define constants for all the possible values of Aisle
const (
	AisleCheeseAndBakery     Aisle = "Cheese & Bakery"
	AisleAlcoholButterCheese Aisle = "18 & 19 (Alcohol, Butter, Cheese)"
	AisleFreezer             Aisle = "16 & 17 (Freezer)"
	AisleNoFoodItems         Aisle = "10-15 (No Food Items)"
	AisleBeveragesAndSnacks  Aisle = "6-9 (Bevs & Snacks)"
	AisleBreakfastAndBaking  Aisle = "3-5 (Breakfast & Baking)"
	AislePastaGlobalCanned   Aisle = "1 & 2 (Pasta, Global, Canned)"
	AisleProduce             Aisle = "Produce"
	AisleMeatAndYogurt       Aisle = "Meat & Yogurt"
)

// AllAisles contains the list of all valid aisle values
var AllAisles = []Aisle{
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

// IsValidAisle checks if the given aisle is a valid enum value
func (a Aisle) IsValid() error {
	for _, validAisle := range AllAisles {
		if a == validAisle {
			return nil
		}
	}
	return errors.New("invalid aisle: " + string(a))
}

type Unit string

const (
	UnitGram  Unit = "gram"
	UnitLb    Unit = "lb"
	UnitOz    Unit = "oz"
	UnitCup   Unit = "cup"
	UnitTbsp  Unit = "tbsp"
	UnitTsp   Unit = "tsp"
	UnitCount Unit = "count"
)

var AllUnits = []Unit{
	UnitGram,
	UnitLb,
	UnitOz,
	UnitCup,
	UnitTbsp,
	UnitTsp,
	UnitCount,
}

func (u Unit) IsValid() error {
	for _, validUnit := range AllUnits {
		if u == validUnit {
			return nil
		}
	}
	return errors.New("invalid unit: " + string(u))
}

type Ingredient struct {
	Name         string  `json:"item"`
	Quantity     float64 `json:"quantity"`
	Unit         Unit    `json:"unit"`
	Aisle        Aisle   `json:"aisle"`
	RelatedMeals []string
}

type Meal struct {
	Name        string       `json:"name"`
	URL         *string      `json:"url,omitempty"`
	Ingredients []Ingredient `json:"ingredients,omitempty"`
}

type Category struct {
	Category string `json:"category"`
	Items    []Meal `json:"items"`
}

func formatSignificant(f float64, sigDigits int) string {
	if f == 0 {
		return "0"
	}

	scale := math.Pow10(sigDigits - 1 - int(math.Floor(math.Log10(math.Abs(f)))))

	result := math.Round(f*scale) / scale

	return fmt.Sprintf("%g", result)
}

func (i Ingredient) String() string {
	return fmt.Sprintf("%s %s: %s (%s)",
		formatSignificant(i.Quantity, 4),   // e.g. "2.75"
		i.Unit,                             // e.g. "lb"
		i.Name,                             // e.g. "Beef"
		strings.Join(i.RelatedMeals, ", "), // e.g. "Burger, Tacos"
	)
}

func MealsToIngredients(meals []Meal) []Ingredient {
	type ingredientKey struct {
		Name  string
		Unit  Unit
		Aisle Aisle
	}

	combined := make(map[ingredientKey]Ingredient)

	// Aggregate ingredients by (Name, Unit, Aisle)
	for _, meal := range meals {
		for _, ing := range meal.Ingredients {
			key := ingredientKey{
				Name:  ing.Name,
				Unit:  ing.Unit,
				Aisle: ing.Aisle,
			}

			agg := combined[key]

			// Update fields and sum quantities
			agg.Name = ing.Name
			agg.Unit = ing.Unit
			agg.Aisle = ing.Aisle
			agg.Quantity += ing.Quantity
			agg.RelatedMeals = append(agg.RelatedMeals, meal.Name)

			combined[key] = agg
		}
	}

	// Flatten into a single slice
	result := make([]Ingredient, 0, len(combined))
	for _, ing := range combined {
		result = append(result, ing)
	}

	return result
}

type MealCollection []Category

// validateIngredient checks if all required fields of an Ingredient are set and valid
func validateIngredient(ingredient Ingredient) error {
	if ingredient.Name == "" {
		return errors.New("ingredient item cannot be empty")
	}
	if ingredient.Quantity <= 0 {
		return errors.New("ingredient quantity must be greater than zero")
	}
	if ingredient.Unit == "" {
		return errors.New("ingredient unit cannot be empty")
	}
	if err := ingredient.Unit.IsValid(); err != nil {
		return err
	}
	if ingredient.Aisle == "" {
		return errors.New("ingredient aisle cannot be empty")
	}
	if err := ingredient.Aisle.IsValid(); err != nil {
		return err
	}
	return nil
}

func validateMealCollection(mealCollection MealCollection) error {
	for _, category := range mealCollection {
		for _, item := range category.Items {
			for _, ingredient := range item.Ingredients {
				if err := validateIngredient(ingredient); err != nil {
					return fmt.Errorf("error in item '%s' of category '%s': %v", item.Name, category.Category, err)
				}
			}
		}
	}
	return nil
}

func OpenMealData(filename string) (io.ReadCloser, error) {
	return os.Open(filename)
}

func OpenFromS3() (io.ReadCloser, error) {
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name is not set")
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	// Create an S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Get the object
	resp, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("recipes.json"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %v", err)
	}

	return resp.Body, nil
}

func ReadMealCollection(reader io.ReadCloser) (MealCollection, error) {
	defer reader.Close()

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	// Decode the JSON data into a MealCollection object
	var mealCollection MealCollection
	if err := decoder.Decode(&mealCollection); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	// Validate the meal collection
	if err := validateMealCollection(mealCollection); err != nil {
		return nil, fmt.Errorf("validation error: %v", err)
	}

	return mealCollection, nil
}

// PopItem removes and returns the first item from the list (if any) and returns the remaining items
func PopItem(items []Meal) (Meal, []Meal, bool) {
	if len(items) == 0 {
		return Meal{}, items, false
	}
	return items[0], items[1:], true
}

func Shuffle(slice interface{}) {
	v := reflect.ValueOf(slice)

	if v.Kind() != reflect.Slice {
		panic("Shuffle() expects a slice")
	}

	rand.Shuffle(v.Len(), func(i, j int) {
		// Swap elements i and j using reflection
		tmp := reflect.ValueOf(v.Index(i).Interface())
		v.Index(i).Set(v.Index(j))
		v.Index(j).Set(tmp)
	})
}

// DeepCopy creates a deep copy of a MealCollection
func (m MealCollection) DeepCopy() MealCollection {
	// Create a new MealCollection
	mealCopy := make(MealCollection, len(m))

	// Copy each Category and its Items
	for i, category := range m {
		// Copy the Category struct (shallow copy)
		mealCopy[i] = category

		// Create a deep copy of the Items slice
		mealCopy[i].Items = make([]Meal, len(category.Items))
		copy(mealCopy[i].Items, category.Items)
	}

	return mealCopy
}

// GenerateMealsWholeYear generates a random list of meals, not respecting categories
func (m MealCollection) GenerateMealsWholeYearNoCategories(currCalendar calendar.Calendar) []Meal {
	// Use Year+Month to make meal generation consistent
	rand.Seed(uint64(currCalendar.Year))

	// Create a copy of MealCollection so that the original isn't modified
	mealCopy := m.DeepCopy()

	var allMeals []Meal
	for _, category := range mealCopy {
		allMeals = append(allMeals, category.Items...)
	}

	currItemInd := 0
	Shuffle(allMeals)

	for i := range int(currCalendar.Month) - 1 {
		pastCalendar := calendar.NewCalendar(currCalendar.Year, time.Month(i+1))

		for j := 1; j < pastCalendar.DaysInMonth()+1; j++ {
			if pastCalendar.GetWeekday(j) == time.Thursday {
				continue
			}
			if pastCalendar.GetWeekday(j) == time.Friday {
				continue
			}

			if currItemInd >= len(allMeals) {
				Shuffle(allMeals)
				currItemInd = 0
			}

			// Copy would go here

			currItemInd += 1
		}
	}

	var selectedMeals []Meal
	for j := 1; j < currCalendar.DaysInMonth()+1; j++ {
		if currCalendar.GetWeekday(j) == time.Thursday {
			selectedMeals = append(selectedMeals, Meal{
				Name: "LEFTOVERS",
			})
			continue
		}
		if currCalendar.GetWeekday(j) == time.Friday {
			selectedMeals = append(selectedMeals, Meal{
				Name: "OUT",
			})
			continue
		}

		if currItemInd >= len(allMeals) {
			Shuffle(allMeals)
			currItemInd = 0
		}

		// Copy would go here
		selectedMeals = append(selectedMeals, allMeals[currItemInd])

		currItemInd += 1
	}

	return selectedMeals
}

// GenerateMealsWholeYear generates a random list of meals by popping one item from each category at a time
func (m MealCollection) GenerateMealsWholeYear(currCalendar calendar.Calendar) []Meal {
	// Use Year+Month to make meal generation consistent
	rand.Seed(uint64(currCalendar.Year))

	// Create a copy of MealCollection so that the original isn't modified
	mealCopy := m.DeepCopy()
	// Shuffle categories
	Shuffle(mealCopy)
	// Shuffle items within each category
	for i := range mealCopy {
		Shuffle(mealCopy[i].Items)
	}

	currMealCategoryIndex := 0
	for i := range int(currCalendar.Month) - 1 {
		pastCalendar := calendar.NewCalendar(currCalendar.Year, time.Month(i+1))

		for j := 1; j < pastCalendar.DaysInMonth()+1; j++ {
			if pastCalendar.GetWeekday(j) == time.Thursday {
				continue
			}
			if pastCalendar.GetWeekday(j) == time.Friday {
				continue
			}

			for {
				// Need to repopulate!
				if currMealCategoryIndex >= len(mealCopy) {
					// Create a copy of MealCollection so that the original isn't modified
					mealCopy = m.DeepCopy()
					// Shuffle categories
					Shuffle(mealCopy)
					// Shuffle items within each category
					for i := range mealCopy {
						Shuffle(mealCopy[i].Items)
					}

					currMealCategoryIndex = 0
				}
				_, remainingItems, popped := PopItem(mealCopy[currMealCategoryIndex].Items)
				if popped {
					mealCopy[currMealCategoryIndex].Items = remainingItems
					currMealCategoryIndex += 1
					break
				}
				currMealCategoryIndex += 1
			}
		}
	}

	var selectedMeals []Meal
	for j := 1; j < currCalendar.DaysInMonth()+1; j++ {
		if currCalendar.GetWeekday(j) == time.Thursday {
			selectedMeals = append(selectedMeals, Meal{
				Name: "LEFTOVERS",
			})
			continue
		}
		if currCalendar.GetWeekday(j) == time.Friday {
			selectedMeals = append(selectedMeals, Meal{
				Name: "OUT",
			})
			continue
		}

		for {
			// Need to repopulate!
			if currMealCategoryIndex >= len(mealCopy) {
				// Create a copy of MealCollection so that the original isn't modified
				mealCopy = m.DeepCopy()
				// Shuffle categories
				Shuffle(mealCopy)
				// Shuffle items within each category
				for i := range mealCopy {
					Shuffle(mealCopy[i].Items)
				}

				currMealCategoryIndex = 0
			}

			item, remainingItems, popped := PopItem(mealCopy[currMealCategoryIndex].Items)
			if popped {
				selectedMeals = append(selectedMeals, item)
				mealCopy[currMealCategoryIndex].Items = remainingItems
				currMealCategoryIndex += 1
				break
			}
			currMealCategoryIndex += 1
		}
	}

	return selectedMeals
}

// GenerateMealsList generates a random list of meals by popping one item from each category at a time
func (m MealCollection) GenerateMealsList(calendar calendar.Calendar) []Meal {
	// Use Year+Month to make meal generation consistent
	rand.Seed(uint64(calendar.Year) + uint64(calendar.Month*10))

	// Create a copy of MealCollection so that the original isn't modified
	mealCopy := m.DeepCopy()

	// Shuffle items within each category
	for i := range mealCopy {
		Shuffle(mealCopy[i].Items)
	}

	var selectedMeals []Meal

	runningDays := 1
	for {
		// Shuffle categories
		Shuffle(mealCopy)

		allPopped := true

		for i := range mealCopy {
			if calendar.GetWeekday(runningDays) == time.Thursday {
				selectedMeals = append(selectedMeals, Meal{
					Name: "LEFTOVERS",
				})
				runningDays += 1
			}

			if calendar.GetWeekday(runningDays) == time.Friday {
				selectedMeals = append(selectedMeals, Meal{
					Name: "OUT",
				})
				runningDays += 1
			}

			if len(mealCopy[i].Items) > 0 {
				item, remainingItems, popped := PopItem(mealCopy[i].Items)
				if popped {
					selectedMeals = append(selectedMeals, item)
					mealCopy[i].Items = remainingItems
					allPopped = false
					runningDays += 1
				}
			}
		}

		if allPopped {
			break
		}
	}

	return selectedMeals
}
