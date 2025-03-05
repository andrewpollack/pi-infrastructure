package meal_backend

import (
	"fmt"
	"meals/calendar"
	"meals/meal_calendar"
	"meals/meal_collection"
	"meals/meal_email"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type DayResponse struct {
	Day     int
	Meal    string
	URL     *string
	Enabled bool
}

type BackendCalendarResponse struct {
	Year          int
	Month         string
	MealsEachWeek [][]DayResponse
}

func CreateBackendCalendarResponse(collection meal_collection.MealCollection, year int, month time.Month) BackendCalendarResponse {
	resp := BackendCalendarResponse{
		Year:          year,
		Month:         month.String(),
		MealsEachWeek: [][]DayResponse{},
	}

	mc := &meal_calendar.MealCalendar{
		Calendar:       *calendar.NewCalendar(year, month),
		MealCollection: collection,
	}

	items := mc.MealCollection.GenerateMealsWholeYearNoCategories(mc.Calendar)

	for _, week := range mc.Calendar.Weeks {
		var weekMeals []DayResponse
		for _, day := range week {
			var item meal_collection.Meal
			switch day.Number {
			case 0:
				item = meal_collection.Meal{
					Name: "",
				}
			default:
				item = items[day.Number-1]
			}

			itemResp := DayResponse{
				Day:     day.Number,
				Meal:    item.Name,
				URL:     item.URL,
				Enabled: !item.Disabled,
			}
			weekMeals = append(weekMeals, itemResp)
		}
		resp.MealsEachWeek = append(resp.MealsEachWeek, weekMeals)
	}

	return resp
}

func getMealCollection(recipeCreatedCutoff int64) (meal_collection.MealCollection, error) {
	collection, err := meal_collection.ReadMealCollectionFromDB(recipeCreatedCutoff)
	if err != nil {
		return nil, err
	}

	return collection, nil
}

func GetCalendar(c *gin.Context) {
	now := time.Now()
	currYear, currMonth, _ := now.Date()
	firstOfMonth := time.Date(currYear, currMonth, 1, 0, 0, 0, 0, now.Location())

	collection, err := getMealCollection(firstOfMonth.Unix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	var nextMonth time.Month
	var nextYear int

	// Determine the next month/year
	if currMonth == time.December {
		nextMonth = time.January
		nextYear = currYear + 1
	} else {
		nextMonth = currMonth + 1
		nextYear = currYear
	}

	currMonthResponse := CreateBackendCalendarResponse(collection, currYear, currMonth)
	nextMonthResponse := CreateBackendCalendarResponse(collection, nextYear, nextMonth)

	c.JSON(http.StatusOK, gin.H{
		"currMonthResponse": currMonthResponse,
		"nextMonthResponse": nextMonthResponse,
	})
}

func GetMeals(c *gin.Context) {
	mealCollection, err := getMealCollection(time.Now().Unix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	// Sort items alphabetically (case-insensitive)
	sort.Slice(mealCollection, func(i, j int) bool {
		return strings.ToLower(mealCollection[i].Name) <
			strings.ToLower(mealCollection[j].Name)
	})

	var allMeals []DayResponse

	for _, item := range mealCollection {
		allMeals = append(allMeals, DayResponse{
			Day:     0,
			Meal:    item.Name,
			URL:     item.URL,
			Enabled: !item.Disabled,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"allMeals": allMeals,
	})
}

func SendEmail(c *gin.Context) {
	var meals []string
	if err := c.BindJSON(&meals); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	// Verify only 5 meals are selected
	if len(meals) != 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Exactly 5 meals must be selected",
		})
		return
	}

	now := time.Now()
	currYear, currMonth, _ := now.Date()
	firstOfMonth := time.Date(currYear, currMonth, 1, 0, 0, 0, 0, now.Location())

	mealCollection, err := getMealCollection(firstOfMonth.Unix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	var currMeals []meal_collection.Meal
	// Attach the meal name to the actual meal collection
	for _, meal := range meals {
		foundItem := false
		for _, item := range mealCollection {
			if item.Name == meal {
				currMeals = append(currMeals, item)
				foundItem = true
				break
			}
		}
		if !foundItem {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Meal not found: %s", meal),
			})
			return
		}
	}

	os.Setenv("H_5", "LEFTOVERS")
	os.Setenv("H_6", "OUT")
	for i, meal := range currMeals {
		if i == 4 {
			os.Setenv("H_7", meal.Name)
			continue
		}
		envVar := fmt.Sprintf("H_%d", i+1)
		os.Setenv(envVar, meal.Name)
	}

	useSES := true
	err = meal_email.CreateAndSendEmail(useSES)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func UpdateMeals(c *gin.Context) {
	var mealUpdates []meal_collection.MealUpdate
	if err := c.BindJSON(&mealUpdates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	mealCollection, err := getMealCollection(time.Now().Unix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	// Filter out only those updates that differ from the current state
	var updatesToApply []meal_collection.MealUpdate
	for _, update := range mealUpdates {
		foundItem := false
		for _, item := range mealCollection {
			if item.Name == update.Name {
				foundItem = true
				// Only include updates if the desired state differs from the current state.
				if item.Disabled != update.Disabled {
					updatesToApply = append(updatesToApply, update)
				}
				break
			}
		}
		if !foundItem {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Meal not found: %s", update.Name),
			})
			return
		}
	}

	// If no updates are needed, return early
	if len(updatesToApply) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status": "no updates needed",
		})
		return
	}

	err = meal_collection.UpdateMealsInDB(updatesToApply)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

func RunBackend() {
	router := gin.Default()

	router.GET("/health", HealthCheck)

	api := router.Group("/api")

	api.GET("/calendar", GetCalendar)
	api.GET("/meals", GetMeals)
	api.POST("/email", SendEmail)
	api.POST("/update", UpdateMeals)

	router.Run()
}
