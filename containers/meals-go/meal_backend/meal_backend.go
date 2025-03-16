package meal_backend

import (
	"fmt"
	"log"
	"meals/calendar"
	"meals/meal_calendar"
	"meals/meal_collection"
	"meals/meal_email"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Config struct {
	PostgresURL    string
	SenderEmail    string
	ReceiverEmails string
}

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

func getMealCollection(postgresURL string, recipeCreatedCutoff int64) (meal_collection.MealCollection, error) {
	collection, err := meal_collection.ReadMealCollectionFromDB(postgresURL, recipeCreatedCutoff)
	if err != nil {
		return nil, err
	}

	return collection, nil
}

func (c Config) GetCalendar(ctx *gin.Context) {
	now := time.Now()
	currYear, currMonth, _ := now.Date()
	firstOfMonth := time.Date(currYear, currMonth, 1, 0, 0, 0, 0, now.Location())

	collection, err := getMealCollection(c.PostgresURL, firstOfMonth.Unix())
	if err != nil {
		log.Println("Error in GetCalendar while fetching meal collection:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
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

	ctx.JSON(http.StatusOK, gin.H{
		"currMonthResponse": currMonthResponse,
		"nextMonthResponse": nextMonthResponse,
	})
}

func (c Config) GetMeals(ctx *gin.Context) {
	mealCollection, err := getMealCollection(c.PostgresURL, time.Now().Unix())
	if err != nil {
		log.Println("Error in GetMeals while fetching meal collection:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
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

	ctx.JSON(http.StatusOK, gin.H{
		"allMeals": allMeals,
	})
}

func (c Config) SendEmail(ctx *gin.Context) {
	var meals []string
	if err := ctx.BindJSON(&meals); err != nil {
		log.Println("Error in SendEmail while binding JSON:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	// Verify only 5 meals are selected
	if len(meals) != 5 {
		errMsg := "Exactly 5 meals must be selected"
		log.Println("Error in SendEmail:", errMsg)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": errMsg,
		})
		return
	}

	mealCollection, err := getMealCollection(c.PostgresURL, time.Now().Unix())
	if err != nil {
		log.Println("Error in SendEmail while fetching meal collection:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	mealMap := mealCollection.MapNameToMeal()
	var currMealNames []string
	for i, meal := range meals {
		// If the meal isn't found in the map, return an error.
		if _, found := mealMap[meal]; !found {
			errMsg := fmt.Sprintf("Meal not found: %s", meal)
			log.Println("Error in SendEmail:", errMsg)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": errMsg,
			})
			return
		}

		// Add Leftovers / Out before the 4th meal.
		if i == 4 {
			currMealNames = append(currMealNames, meal_collection.MEAL_LEFTOVERS.Name, meal_collection.MEAL_OUT.Name)
		}
		currMealNames = append(currMealNames, meal)
	}

	mealEmailConfig := meal_email.Config{
		PostgresURL:    c.PostgresURL,
		UseSES:         true,
		HardcodedMeals: currMealNames,
		SenderEmail:    c.SenderEmail,
		ReceiverEmails: c.ReceiverEmails,
	}
	err = mealEmailConfig.CreateAndSendEmail()
	if err != nil {
		log.Println("Error in SendEmail while creating and sending email:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func (c Config) DisableMeals(ctx *gin.Context) {
	var mealUpdates []meal_collection.MealUpdate
	if err := ctx.BindJSON(&mealUpdates); err != nil {
		log.Println("Error in DisableMeals while binding JSON:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	mealCollection, err := getMealCollection(c.PostgresURL, time.Now().Unix())
	if err != nil {
		log.Println("Error in DisableMeals while fetching meal collection:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	mealMap := mealCollection.MapNameToMeal()

	var updatesToApply []meal_collection.MealUpdate
	for _, update := range mealUpdates {
		meal, ok := mealMap[update.Name]
		if !ok {
			errMsg := fmt.Sprintf("Meal not found: %s", update.Name)
			log.Println("Error in DisableMeals:", errMsg)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": errMsg,
			})
			return
		}
		// Only include updates if the desired state differs from the current state.
		if meal.Disabled != update.Disabled {
			updatesToApply = append(updatesToApply, update)
		}
	}

	// If no updates are needed, return early
	if len(updatesToApply) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "no updates needed",
		})
		return
	}

	err = meal_collection.UpdateMealsInDB(c.PostgresURL, updatesToApply)
	if err != nil {
		log.Println("Error in DisableMeals while updating meals in DB:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

func (c Config) RunBackend() {
	router := gin.Default()

	router.GET("/health", HealthCheck)

	api := router.Group("/api")

	api.GET("/calendar", c.GetCalendar)
	api.GET("/meals", c.GetMeals)
	api.POST("/email", c.SendEmail)
	api.POST("/update", c.DisableMeals)

	router.Run()
}
