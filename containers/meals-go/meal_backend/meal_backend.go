package meal_backend

import (
	"meals/calendar"
	"meals/meal_calendar"
	"meals/meal_collection"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type DayResponse struct {
	Day  int
	Meal string
	URL  *string
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
				Day:  day.Number,
				Meal: item.Name,
				URL:  item.URL,
			}
			weekMeals = append(weekMeals, itemResp)
		}
		resp.MealsEachWeek = append(resp.MealsEachWeek, weekMeals)
	}

	return resp
}

func getMealCollection() (meal_collection.MealCollection, error) {
	collection, err := meal_collection.ReadMealCollectionFromDB()
	if err != nil {
		return nil, err
	}

	return collection, nil
}

func GetCalendar(c *gin.Context) {
	collection, err := getMealCollection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	currYear, currMonth, _ := time.Now().Date()
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
	collection, err := getMealCollection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	var flattenedMeals []meal_collection.Meal
	for _, item := range collection {
		flattenedMeals = append(flattenedMeals, item.Items...)
	}

	// Sort items alphabetically (case-insensitive)
	sort.Slice(flattenedMeals, func(i, j int) bool {
		return strings.ToLower(flattenedMeals[i].Name) <
			strings.ToLower(flattenedMeals[j].Name)
	})

	var allMeals []DayResponse

	for _, item := range flattenedMeals {
		allMeals = append(allMeals, DayResponse{
			Day:  0,
			Meal: item.Name,
			URL:  item.URL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"allMeals": allMeals,
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

	api.GET("/calendar", GetCalendar) // api: /api/ping
	api.GET("/meals", GetMeals)       // api: /api/ping

	router.Run()
}
