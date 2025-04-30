package meal_backend

import (
	"fmt"
	"log"
	"meals/calendar"
	"meals/config"
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

// Config holds the configuration for the backend.
type Config struct {
	PostgresURL          string
	PostgresMigrationDir string
	EmailSender          string
	EmailReceivers       []string
	AllowOrigins         []string
	DomainName           string
	JWTSigningKey        []byte
	DeploymentPassword   string
}

// DayResponse represents a meal for a given day.
type DayResponse struct {
	Day     int
	Meal    string
	URL     *string
	Enabled bool
}

// ExtraItemResponse represents an extra item response.
type ExtraItemResponse struct {
	Items []meal_collection.ExtraItem
}

// BackendCalendarResponse represents the calendar response.
type BackendCalendarResponse struct {
	Year          int
	Month         string
	MealsEachWeek [][]DayResponse
}

// CreateBackendCalendarResponse creates a calendar response.
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

// GetCalendar handles the GET /calendar endpoint.
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

// GetMeals handles the GET /meals endpoint.
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

	allMeals := make([]DayResponse, 0, len(mealCollection))
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

// GetItems handles the GET /items endpoint.
func (c Config) GetItems(ctx *gin.Context) {
	type ExtraItemResponse struct {
		Name    string `json:"Name"`
		Aisle   string `json:"Aisle"`
		ID      int    `json:"ID"`
		Enabled bool   `json:"Enabled"`
	}

	extraItems, err := meal_collection.ReadExtraItemsFromDB(c.PostgresURL)
	if err != nil {
		log.Println("Error in GetItems while fetching extra items:", err)
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

	extraItemsResponse := make([]ExtraItemResponse, 0, len(extraItems))
	for _, item := range extraItems {
		extraItemsResponse = append(extraItemsResponse, ExtraItemResponse{
			Name:    item.Name,
			Aisle:   string(item.Aisle),
			ID:      item.ID,
			Enabled: item.Enabled,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"allItems": extraItemsResponse,
	})
}

// SendEmailRequest represents the email request payload.
type SendEmailRequest struct {
	Meals      []string `json:"meals"`
	Emails     []string `json:"emails"`
	ExtraItems []string `json:"extraItems"`
}

// SendEmail handles the POST /email endpoint.
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
	// Verify emails are not empty, and are a subset of the allowed emails in c.EmailReceivers
	if len(emails) == 0 {
		errMsg := "At least one email must be provided"
		log.Println("Error in SendEmail:", errMsg)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": errMsg,
		})
		return
	}
	allowedEmails := make(map[string]bool)
	for _, email := range c.EmailReceivers {
		allowedEmails[email] = true
	}
	for _, email := range emails {
		if _, found := allowedEmails[email]; !found {
			errMsg := fmt.Sprintf("Email not allowed: %s", email)
			log.Println("Error in SendEmail:", errMsg)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": errMsg,
			})
			return
		}
	}

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
		log.Println("Error in SendEmail while fetching extra items:", err)
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
		Receivers:      emails,
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

// EnableMeals handles the POST /meals/enable endpoint.
func (c Config) EnableMeals(ctx *gin.Context) {
	var mealUpdates []meal_collection.MealUpdate
	if err := ctx.BindJSON(&mealUpdates); err != nil {
		log.Println("Error in EnableMeals while binding JSON:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	mealCollection, err := meal_collection.ReadMealCollectionFromDB(c.PostgresURL, time.Now().Unix())
	if err != nil {
		log.Println("Error in EnableMeals while fetching meal collection:", err)
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
			log.Println("Error in EnableMeals:", errMsg)
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
		log.Println("Error in EnableMeals while updating meals in DB:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

// UpdateItems handles the POST /items/update endpoint.
func (c Config) UpdateItems(ctx *gin.Context) {
	var extraItemsUpdate []meal_collection.FEExtraItem
	if err := ctx.BindJSON(&extraItemsUpdate); err != nil {
		log.Println("Error in UpdateItems while binding JSON:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	err := meal_collection.UpdateExtraItemsInDB(c.PostgresURL, extraItemsUpdate)
	if err != nil {
		log.Println("Error in UpdateItems while updating items in DB:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

// HealthCheck handles the GET /health endpoint.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

// PostLoginRequest represents the login request payload.
type PostLoginRequest struct {
	Password string `json:"password"`
}

// Login handles the POST /login endpoint.
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

// Auth verifies the authentication token.
func (c Config) Auth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

// GetAisles handles the GET /aisles endpoint to return configured Aisles.
func (c Config) GetAisles(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"aisles": config.Cfg.App.Aisles,
	})
}

// GetEmails handles the GET /emails endpoint to return configured email receivers.
func (c Config) GetEmails(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"emails": c.EmailReceivers,
	})
}

// RunBackend initializes migrations and starts the Gin router.
func (c Config) RunBackend() {
	// TODO: At some point, it would be nice to run migrations not in this
	// server, but in a separate one. This would allow us to run multiple
	// copies of the backend without worrying about migration conflicts.
	c.runMigrations()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     c.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	router.GET("/health", HealthCheck)
	router.GET("/auth", c.authenticateMiddleware, c.Auth)

	api := router.Group("/api")
	api.POST("/login", c.Login)

	// Require authentication for all other routes
	api.GET("/calendar", c.authenticateMiddleware, c.GetCalendar)
	api.GET("/items", c.authenticateMiddleware, c.GetItems)
	api.POST("/items/update", c.authenticateMiddleware, c.UpdateItems)
	api.POST("/email", c.authenticateMiddleware, c.SendEmail)
	api.GET("/meals", c.authenticateMiddleware, c.GetMeals)
	api.POST("/meals/enable", c.authenticateMiddleware, c.EnableMeals)
	api.GET("/aisles", c.authenticateMiddleware, c.GetAisles)
	api.GET("/emails", c.authenticateMiddleware, c.GetEmails)

	err := router.Run()
	if err != nil {
		log.Fatal(err)
	}
}
