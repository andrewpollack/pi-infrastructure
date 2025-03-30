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
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Config struct {
	PostgresURL        string
	EmailSender        string
	EmailReceivers     string
	IgnoreCutoff       bool
	AllowOrigins       []string
	DomainName         string
	JWTSigningKey      []byte
	DeploymentPassword string
}

type DayResponse struct {
	Day     int
	Meal    string
	URL     *string
	Enabled bool
}

type ExtraItemResponse struct {
	Items []meal_collection.ExtraItem
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

func (c Config) GetCalendar(ctx *gin.Context) {
	now := time.Now()
	currYear, currMonth, _ := now.Date()

	// Get optional query parameters
	yearStr := ctx.Query("year")
	monthStr := ctx.Query("month")

	var year int
	var month time.Month
	var err error
	if yearStr != "" && monthStr != "" {
		// Parse the query parameters
		y, err := strconv.Atoi(yearStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year parameter"})
			return
		}
		m, err := strconv.Atoi(monthStr)
		if err != nil || m < 1 || m > 12 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month parameter"})
			return
		}
		year = y
		month = time.Month(m)
	} else {
		year = currYear
		month = currMonth
	}

	collection, err := meal_collection.ReadMealCollectionFromDB(c.PostgresURL, now.Unix())
	if err != nil {
		log.Println("Error in GetCalendar while fetching meal collection:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	monthResponse := CreateBackendCalendarResponse(collection, year, month)

	ctx.JSON(http.StatusOK, gin.H{
		"currMonthResponse": monthResponse,
	})
}

func (c Config) GetMeals(ctx *gin.Context) {
	mealCollection, err := meal_collection.ReadMealCollectionFromDB(c.PostgresURL, time.Now().Unix())
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

func (c Config) GetItems(ctx *gin.Context) {
	type ExtraItemResponse struct {
		Name string `json:"Name"`
	}

	extraItems, err := meal_collection.ReadExtraItemsFromDB(c.PostgresURL)
	if err != nil {
		log.Println("Error in GetMeals while fetching meal collection:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	// Sort items alphabetically (case-insensitive)
	sort.Slice(extraItems, func(i, j int) bool {
		return strings.ToLower(extraItems[i].Name) <
			strings.ToLower(extraItems[j].Name)
	})

	var extraItemsResponse []ExtraItemResponse
	for _, item := range extraItems {
		extraItemsResponse = append(extraItemsResponse, ExtraItemResponse{
			Name: item.Name,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"allItems": extraItemsResponse,
	})
}

type SendEmailRequest struct {
	Meals      []string `json:"meals"`
	Emails     []string `json:"emails"`
	ExtraItems []string `json:"extraItems"`
}

func (c Config) SendEmail(ctx *gin.Context) {
	var emailRequest SendEmailRequest
	if err := ctx.BindJSON(&emailRequest); err != nil {
		log.Println("Error in SendEmail while binding JSON:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	meals := emailRequest.Meals
	emails := emailRequest.Emails
	extraItems := emailRequest.ExtraItems

	if len(emails) == 0 {
		errMsg := "At least one email must be provided"
		log.Println("Error in SendEmail:", errMsg)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": errMsg,
		})
		return
	}

	// Verify 7 meals are selected
	if len(meals) != 7 {
		errMsg := "Exactly 7 meals must be selected"
		log.Println("Error in SendEmail:", errMsg)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": errMsg,
		})
		return
	}

	mealCollection, err := meal_collection.ReadMealCollectionFromDB(c.PostgresURL, time.Now().Unix())
	if err != nil {
		log.Println("Error in SendEmail while fetching meal collection:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	extraItemsDB, err := meal_collection.ReadExtraItemsFromDB(c.PostgresURL)
	if err != nil {
		log.Println("Error in GetMeals while fetching meal collection:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	mealMap := mealCollection.MapNameToMeal()
	var currMealNames []string
	for _, meal := range meals {
		// If the meal isn't found in the map, return an error.
		if _, found := mealMap[meal]; !found {
			errMsg := fmt.Sprintf("Meal not found: %s", meal)
			log.Println("Error in SendEmail:", errMsg)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": errMsg,
			})
			return
		}

		currMealNames = append(currMealNames, meal)
	}

	extraItemsMap := map[string]meal_collection.ExtraItem{}
	for _, extraItem := range extraItemsDB {
		extraItemsMap[extraItem.Name] = extraItem
	}
	var extraItemNames []string
	for _, extraItem := range extraItems {
		// If the extraItem isn't found in the map, return an error.
		if _, found := extraItemsMap[extraItem]; !found {
			errMsg := fmt.Sprintf("Extra Item not found: %s", extraItem)
			log.Println("Error in SendEmail:", errMsg)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": errMsg,
			})
			return
		}

		extraItemNames = append(extraItemNames, extraItem)
	}

	mealEmailConfig := meal_email.Config{
		PostgresURL:    c.PostgresURL,
		EmailService:   meal_email.SES,
		HardcodedMeals: currMealNames,
		Sender:         c.EmailSender,
		Receivers:      strings.Join(emails, ","),
		ExtraItems:     extraItemNames,
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

	mealCollection, err := meal_collection.ReadMealCollectionFromDB(c.PostgresURL, time.Now().Unix())
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

type PostLoginRequest struct {
	Password string `json:"password"`
}

func (c Config) Login(ctx *gin.Context) {
	var loginRequest PostLoginRequest
	if err := ctx.BindJSON(&loginRequest); err != nil {
		log.Println("Error in Login while binding JSON:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	if loginRequest.Password == c.DeploymentPassword {
		tokenString, err := createToken(c.JWTSigningKey)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
			return
		}
		ctx.SetCookie("token", tokenString, 2.628e+6, "/", "", false, true)
		ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
	} else {
		time.Sleep(2 * time.Second)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}

func (c Config) RunBackend() {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     c.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	router.GET("/health", HealthCheck)

	api := router.Group("/api")
	api.POST("/login", c.Login)
	api.GET("/calendar", c.authenticateMiddleware, c.GetCalendar)
	api.GET("/meals", c.authenticateMiddleware, c.GetMeals)
	api.GET("/items", c.authenticateMiddleware, c.GetItems)
	api.POST("/email", c.authenticateMiddleware, c.SendEmail)
	api.POST("/update", c.authenticateMiddleware, c.DisableMeals)

	err := router.Run()
	if err != nil {
		log.Fatal(err)
	}
}
