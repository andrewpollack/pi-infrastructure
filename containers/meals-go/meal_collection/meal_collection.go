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
	"sort"
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

func (u Unit) IsValid() error {
	var validUnits = map[Unit]struct{}{
		UnitGram:  {},
		UnitLb:    {},
		UnitOz:    {},
		UnitCup:   {},
		UnitTbsp:  {},
		UnitTsp:   {},
		UnitCount: {},
	}

	if _, ok := validUnits[u]; ok {
		return nil
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
	Disabled    bool         `json:"disabled,omitempty"`
	Category    *string      `json:"category,omitempty"`
}

var MEAL_LEFTOVERS = Meal{
	Name: "Leftovers",
}

var MEAL_OUT = Meal{
	Name: "Out",
}

type MealCollection []Meal

// TODO: Name this something more logical... maybe just "Item"?
type ExtraItem struct {
	Name    string `json:"name"`
	Aisle   Aisle  `json:"aisle"`
	Enabled bool   `json:"enabled"`
	ID      int    `json:"id"`
}

func (m MealCollection) MapNameToMeal() map[string]Meal {
	mealMap := map[string]Meal{
		MEAL_LEFTOVERS.Name: MEAL_LEFTOVERS,
		MEAL_OUT.Name:       MEAL_OUT,
	}

	for _, meal := range m {
		mealMap[meal.Name] = meal
	}

	return mealMap
}

func (i Ingredient) String() string {
	if i.Quantity == 0 {
		return fmt.Sprint(i.Name)
	}
	formatSignificant := func(f float64, sigDigits int) string {
		if f == 0 {
			return "0"
		}
		scale := math.Pow10(sigDigits - 1 - int(math.Floor(math.Log10(math.Abs(f)))))
		result := math.Round(f*scale) / scale
		return fmt.Sprintf("%g", result)
	}

	return fmt.Sprintf("%s %s: %s (%s)",
		formatSignificant(i.Quantity, 4),   // e.g. "2.75"
		i.Unit,                             // e.g. "lb"
		i.Name,                             // e.g. "Beef"
		strings.Join(i.RelatedMeals, ", "), // e.g. "Burger, Tacos"
	)
}

func (i Ingredient) StringBolded() string {
	if i.Quantity == 0 {
		return fmt.Sprintf("<strong>%s</strong>", i.Name)
	}
	formatSignificant := func(f float64, sigDigits int) string {
		if f == 0 {
			return "0"
		}
		scale := math.Pow10(sigDigits - 1 - int(math.Floor(math.Log10(math.Abs(f)))))
		result := math.Round(f*scale) / scale
		return fmt.Sprintf("%g", result)
	}

	// Extract the final word of each item in RelatedMeals
	mealWords := make([]string, len(i.RelatedMeals))
	for idx, meal := range i.RelatedMeals {
		words := strings.Fields(meal)
		if len(words) > 0 {
			mealWords[idx] = words[len(words)-1]
		} else {
			mealWords[idx] = meal
		}
	}

	return fmt.Sprintf("<strong>%s</strong> - %s %s (%s)",
		i.Name,                           // e.g. "Beef"
		formatSignificant(i.Quantity, 4), // e.g. "2.75"
		i.Unit,                           // e.g. "lb"
		strings.Join(mealWords, ", "),    // e.g. "Burger, Tacos"
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
			sort.Strings(agg.RelatedMeals)

			combined[key] = agg
		}
	}

	// Flatten into a single slice
	result := make([]Ingredient, 0, len(combined))
	for _, ing := range combined {
		result = append(result, ing)
	}

	// Sort by agg.RelatedMeals first, then by Name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	// Sort by agg.RelatedMeals
	sort.Slice(result, func(i, j int) bool {
		return strings.Join(result[i].RelatedMeals, ", ") < strings.Join(result[j].RelatedMeals, ", ")
	})

	return result
}

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
	for _, item := range mealCollection {
		for _, ingredient := range item.Ingredients {
			if err := validateIngredient(ingredient); err != nil {
				return fmt.Errorf("error in item '%s' of category '%s': %v", item.Name, *item.Category, err)
			}
		}
	}
	return nil
}

func OpenMealData(filename string) (io.ReadCloser, error) {
	return os.Open(filename)
}

func OpenFromS3(bucketName string, bucketKey string) (io.ReadCloser, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name is not set")
	}
	if bucketKey == "" {
		return nil, fmt.Errorf("bucket key is not set")
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
		Key:    aws.String(bucketKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %v", err)
	}

	return resp.Body, nil
}

func ExtraItemToIngredient(ei ExtraItem) Ingredient {
	return Ingredient{
		Name:     ei.Name,
		Aisle:    ei.Aisle,
		Quantity: 0,
	}
}

func ReadMealCollectionFromReader(reader io.ReadCloser) (MealCollection, error) {
	defer func() {
		if err := reader.Close(); err != nil {
			fmt.Printf("error closing reader: %v\n", err)
		}
	}()

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	// Decode the JSON data into a MealCollection object
	var mealCollection MealCollection
	if err := decoder.Decode(&mealCollection); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	// Sort meals by name
	sort.Slice(mealCollection, func(i, j int) bool {
		return strings.ToLower(mealCollection[i].Name) < strings.ToLower(mealCollection[j].Name)
	})

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

	// Copy each Meal
	for i, category := range m {
		mealCopy[i] = category
		mealCopy[i].Ingredients = make([]Ingredient, len(category.Ingredients))
		copy(mealCopy[i].Ingredients, category.Ingredients)
	}

	return mealCopy
}

// GenerateMealsWholeYear generates a random list of meals, not respecting categories, and
// starting from the beginning of the year
func (m MealCollection) GenerateMealsWholeYearNoCategories(currCalendar calendar.Calendar) []Meal {
	// Use Year to make meal generation consistent
	rand.Seed(uint64(currCalendar.Year))

	// Check if the target month is before or after the current calendar month
	futureMonth := time.Now().Month() <= currCalendar.Month

	// Create a copy of MealCollection so that the original isn't modified
	mealCopy := m.DeepCopy()

	var allMeals []Meal
	for _, item := range mealCopy {
		allMeals = append(allMeals, item)
	}

	currItemInd := 0
	totalShuffles := 0
	Shuffle(allMeals)
	appendItems := false
	var selectedMeals []Meal
	for i := 1; i <= int(currCalendar.Month); i++ {
		// Keep cycling through items and shuffling until we get to the desired month
		if time.Month(i) == currCalendar.Month {
			appendItems = true
		}

		cal := calendar.NewCalendar(currCalendar.Year, time.Month(i))
		for j := 1; j <= cal.DaysInMonth(); j++ {
			startingShuffleNum := totalShuffles
			var item Meal
			switch cal.GetWeekday(j) {
			case time.Thursday:
				item = MEAL_LEFTOVERS
			case time.Friday:
				item = MEAL_OUT
			default:
				if appendItems || futureMonth {
					// Skip until finding a meal that is enabled. Makes shuffling more consistent
					// month to month for previous months.
					for {
						if allMeals[currItemInd].Disabled {
							currItemInd += 1
							if currItemInd >= len(allMeals) {
								Shuffle(allMeals)
								currItemInd = 0
								totalShuffles += 1
							}
						} else {
							break
						}
					}
				}

				item = allMeals[currItemInd]

				currItemInd += 1
				if currItemInd >= len(allMeals) {
					Shuffle(allMeals)
					currItemInd = 0
					totalShuffles += 1
				}
				if totalShuffles > startingShuffleNum {
					item.Name = fmt.Sprintf("%s**", item.Name)
				}
			}

			if appendItems {
				selectedMeals = append(selectedMeals, item)
			}
		}
	}

	return selectedMeals
}
