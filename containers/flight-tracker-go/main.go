package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type FlightInfo struct {
	FlightNumber   string
	Departure      string
	Arrival        string
	DepartureTime  string
	ArrivalTime    string
	AircraftType   string
	Status         string
	Owner          string
	OwnerLocation  string
	DirectDistance string
	FlightDuration string
	EstimatedFuel  string
}

// Flight data from FlightAware JSON
type FlightAwareData struct {
	Flights []struct {
		Origin struct {
			FriendlyName     string `json:"friendlyName"`
			FriendlyLocation string `json:"friendlyLocation"`
			Icao             string `json:"icao"`
		} `json:"origin"`
		Destination struct {
			FriendlyName     string `json:"friendlyName"`
			FriendlyLocation string `json:"friendlyLocation"`
			Icao             string `json:"icao"`
		} `json:"destination"`
		AircraftType         string `json:"aircraftType"`
		AircraftTypeFriendly string `json:"aircraftTypeFriendly"`
		TakeoffTimes         struct {
			Actual int64 `json:"actual"`
		} `json:"takeoffTimes"`
		LandingTimes struct {
			Actual int64 `json:"actual"`
		} `json:"landingTimes"`
		FlightStatus string `json:"flightStatus"`
		FlightPlan   struct {
			DirectDistance int `json:"directDistance"`
			Ete            int `json:"ete"`
			FuelBurn       struct {
				Gallons int `json:"gallons"`
				Pounds  int `json:"pounds"`
			} `json:"fuelBurn"`
		} `json:"flightPlan"`
		Aircraft struct {
			Tail          string `json:"tail"`
			Owner         string `json:"owner"`
			OwnerLocation string `json:"ownerLocation"`
			FriendlyType  string `json:"friendlyType"`
		} `json:"aircraft"`
	} `json:"flights"`
}

func fetchFlightInfo(tailNumber string) (*FlightInfo, error) {
	url := fmt.Sprintf("https://www.flightaware.com/live/flight/%s", tailNumber)
	fmt.Println("Fetching flight information for", tailNumber, "from ", url)

	// Create a custom HTTP client to mimic a browser
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers to appear more like a browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch flight page: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch flight info: status code %d", resp.StatusCode)
	}

	// Read the entire response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Convert body to string
	htmlContent := string(body)

	// Extract flight data from JavaScript object in the HTML
	// First, find the trackpollBootstrap variable
	trackpollRegex := regexp.MustCompile(`var\s+trackpollBootstrap\s*=\s*({[\s\S]*?});`)
	trackpollMatches := trackpollRegex.FindStringSubmatch(htmlContent)
	if len(trackpollMatches) < 2 {
		return nil, fmt.Errorf("could not find trackpollBootstrap data in the response")
	}

	// Now extract the activityLog.flights array from the JSON
	trackpollData := trackpollMatches[1]

	// Look for the flights array in the activityLog
	flightDataRegex := regexp.MustCompile(`"activityLog"\s*:\s*{\s*"flights"\s*:\s*(\[\s*{[\s\S]*?\}\s*\])`)
	matches := flightDataRegex.FindStringSubmatch(trackpollData)
	if len(matches) < 2 {
		// Try an alternative pattern if the first one fails
		altRegex := regexp.MustCompile(`flights"\s*:\s*(\[\s*{[\s\S]*?\}\s*\])`)
		altMatches := altRegex.FindStringSubmatch(trackpollData)
		if len(altMatches) < 2 {
			// Save the HTML for debugging
			err := os.WriteFile("debug_response.html", []byte(htmlContent), 0644)
			if err != nil {
				fmt.Printf("Warning: Failed to save debug file: %v\n", err)
			}
			return nil, fmt.Errorf("could not find flight data in the response (saved debug_response.html for inspection)")
		}
		matches = altMatches
	}

	// Parse the JSON data
	var flightData FlightAwareData
	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"flights":%s}`, matches[1])), &flightData); err != nil {
		return nil, fmt.Errorf("failed to parse flight data: %v", err)
	}

	// Check if any flights were found
	if len(flightData.Flights) == 0 {
		return nil, fmt.Errorf("no flight information found for tail number %s", tailNumber)
	}

	// Get the most recent flight (first in the list)
	latestFlight := flightData.Flights[0]

	// Format times
	departureTime := time.Unix(latestFlight.TakeoffTimes.Actual, 0).Format("2006-01-02 15:04:05 MST")
	arrivalTime := time.Unix(latestFlight.LandingTimes.Actual, 0).Format("2006-01-02 15:04:05 MST")

	// Calculate flight duration
	duration := time.Duration(latestFlight.FlightPlan.Ete) * time.Second
	durationStr := formatDuration(duration)

	// Construct FlightInfo
	flightInfo := &FlightInfo{
		FlightNumber:   tailNumber,
		Departure:      fmt.Sprintf("%s (%s)", latestFlight.Origin.FriendlyName, latestFlight.Origin.FriendlyLocation),
		Arrival:        fmt.Sprintf("%s (%s)", latestFlight.Destination.FriendlyName, latestFlight.Destination.FriendlyLocation),
		DepartureTime:  departureTime,
		ArrivalTime:    arrivalTime,
		AircraftType:   latestFlight.AircraftTypeFriendly,
		Status:         toTitleCase(latestFlight.FlightStatus),
		Owner:          latestFlight.Aircraft.Owner,
		OwnerLocation:  latestFlight.Aircraft.OwnerLocation,
		DirectDistance: fmt.Sprintf("%d miles", latestFlight.FlightPlan.DirectDistance),
		FlightDuration: durationStr,
		EstimatedFuel:  fmt.Sprintf("%d gallons (%d pounds)", latestFlight.FlightPlan.FuelBurn.Gallons, latestFlight.FlightPlan.FuelBurn.Pounds),
	}

	return flightInfo, nil
}

// Helper function to format duration in a readable way
func formatDuration(d time.Duration) string {
	totalMinutes := int(d.Minutes())
	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	if hours > 0 {
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}
	return fmt.Sprintf("%d minutes", minutes)
}

// toTitleCase capitalizes the first letter of each word in a string
// This is a replacement for the deprecated strings.Title function
func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

// Define Prometheus metrics
var (
	flightInfoRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flight_tracker_requests_total",
		Help: "The total number of flight information requests",
	})

	flightInfoErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flight_tracker_errors_total",
		Help: "The total number of flight information request errors",
	})

	flightInfoDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "flight_tracker_request_duration_seconds",
		Help:    "The duration of flight information requests in seconds",
		Buckets: prometheus.DefBuckets,
	})

	flightStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flight_tracker_status",
		Help: "Current flight status (1=in air, 0=landed/not flying)",
	}, []string{"tail_number", "owner"})

	flightDistance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flight_tracker_distance_miles",
		Help: "Flight distance in miles",
	}, []string{"tail_number"})

	flightFuel = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flight_tracker_fuel_gallons",
		Help: "Estimated fuel usage in gallons",
	}, []string{"tail_number"})
)

// updateFlightMetrics updates the Prometheus metrics with flight information
func updateFlightMetrics(flightInfo *FlightInfo) {
	// Update flight status metric (1 if in air, 0 if landed/not flying)
	statusValue := 0.0
	if strings.Contains(strings.ToLower(flightInfo.Status), "air") {
		statusValue = 1.0
	}
	flightStatus.WithLabelValues(flightInfo.FlightNumber, flightInfo.Owner).Set(statusValue)

	// Extract numeric values from string fields
	// Parse distance (format: "X miles")
	var distance float64
	_, err := fmt.Sscanf(flightInfo.DirectDistance, "%f miles", &distance)
	if err != nil {
		log.Printf("Error parsing distance '%s': %v", flightInfo.DirectDistance, err)
		// Try to extract using a different approach if the format doesn't match
		distanceStr := strings.TrimSuffix(strings.TrimSpace(flightInfo.DirectDistance), " miles")
		if parsedDistance, parseErr := strconv.ParseFloat(distanceStr, 64); parseErr == nil {
			distance = parsedDistance
		}
	}
	flightDistance.WithLabelValues(flightInfo.FlightNumber).Set(distance)

	// Parse fuel (format: "X gallons (Y pounds)")
	var fuelGallons float64
	_, err = fmt.Sscanf(flightInfo.EstimatedFuel, "%f gallons", &fuelGallons)
	if err != nil {
		log.Printf("Error parsing fuel '%s': %v", flightInfo.EstimatedFuel, err)
		// Try to extract using a different approach if the format doesn't match
		parts := strings.Split(flightInfo.EstimatedFuel, " gallons")
		if len(parts) > 0 {
			if parsedFuel, parseErr := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); parseErr == nil {
				fuelGallons = parsedFuel
			}
		}
	}
	flightFuel.WithLabelValues(flightInfo.FlightNumber).Set(fuelGallons)
}

// fetchFlightInfoWithMetrics wraps fetchFlightInfo with metrics collection
func fetchFlightInfoWithMetrics(tailNumber string) (*FlightInfo, error) {
	flightInfoRequests.Inc()
	startTime := time.Now()

	flightInfo, err := fetchFlightInfo(tailNumber)
	if err != nil {
		flightInfoErrors.Inc()
		return nil, err
	}

	duration := time.Since(startTime).Seconds()
	flightInfoDuration.Observe(duration)

	// Update metrics with flight information
	updateFlightMetrics(flightInfo)

	return flightInfo, nil
}

// serveFlightInfo handles HTTP requests for flight information
func serveFlightInfo(w http.ResponseWriter, r *http.Request) {
	tailNumber := r.URL.Query().Get("tail")

	flightInfo, err := fetchFlightInfoWithMetrics(tailNumber)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching flight info: %v", err), http.StatusInternalServerError)
		return
	}

	// Return flight info as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(flightInfo); err != nil {
		log.Printf("Error encoding flight info to JSON: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// updateAllFlightData updates flight data for all tail numbers
func updateAllFlightData(tailNumbers []string) {
	// Update flight data for each tail number
	for _, tailNumber := range tailNumbers {
		log.Printf("Updating flight data for %s", tailNumber)
		flightInfo, err := fetchFlightInfoWithMetrics(tailNumber)
		if err != nil {
			log.Printf("Error updating flight data for %s: %v", tailNumber, err)
			// When there's an error, explicitly set the flight status to 0 (not flying/unknown)
			// This ensures the metric is updated even when we can't fetch the flight data
			flightStatus.WithLabelValues(tailNumber, "UNKNOWN").Set(0.0)
			continue
		}
		log.Printf("Successfully updated flight data for %s: Status=%s", 
			tailNumber, flightInfo.Status)
		
		// Log whether the flight is in the air or not for monitoring
		if strings.Contains(strings.ToLower(flightInfo.Status), "air") {
			// Explicitly update the flight status metric to 1.0 (in air)
			flightStatus.WithLabelValues(tailNumber, flightInfo.Owner).Set(1.0)
			log.Printf("%s is currently in the air", tailNumber)
		} else {
			// Explicitly update the flight status metric to 0.0 (not in air)
			flightStatus.WithLabelValues(tailNumber, flightInfo.Owner).Set(0.0)
			log.Printf("%s is not currently in the air (status: %s)", tailNumber, flightInfo.Status)
		}
	}
}

// scheduleFlightDataUpdates starts a background goroutine that periodically
// fetches flight data to keep metrics updated for Prometheus scraping
func scheduleFlightDataUpdates(tailNumbers []string, baseInterval time.Duration, maxJitter time.Duration) {
	// Run an immediate update at program start
	log.Printf("Running immediate flight data update at startup")
	updateAllFlightData(tailNumbers)
	
	go func() {
		for {
			// Add random jitter to the interval (0 to maxJitter)
			jitter := time.Duration(int64(float64(maxJitter) * rand.Float64()))
			interval := baseInterval + jitter

			// Log the next update time
			nextUpdate := time.Now().Add(interval)
			log.Printf("Next flight data update scheduled at %s (in %s)",
				nextUpdate.Format("2006-01-02 15:04:05"),
				interval.Round(time.Second))

			// Wait for the interval
			time.Sleep(interval)

			// Update all flight data
			updateAllFlightData(tailNumbers)
		}
	}()
}

func main() {
	// No need to seed the random number generator in Go 1.20+
	// The math/rand package is automatically seeded now

	// Get tail numbers from environment variable
	envTails := os.Getenv("TAIL_NUMBERS")
	if envTails == "" {
		log.Fatal("TAIL_NUMBERS environment variable must be set with comma-separated tail numbers to track")
	}

	// Parse comma-separated tail numbers
	tailNumbers := strings.Split(envTails, ",")

	// Validate tail numbers
	if len(tailNumbers) == 0 {
		log.Fatal("No tail numbers provided in TAIL_NUMBERS environment variable")
	}

	// Trim whitespace from each tail number
	for i, tail := range tailNumbers {
		tailNumber := strings.TrimSpace(tail)
		if tailNumber == "" {
			log.Fatal("Empty tail number found in TAIL_NUMBERS environment variable")
		}
		tailNumbers[i] = tailNumber
	}

	// Schedule background updates for flight data
	// Update every 5 minutes with up to 60 seconds of random jitter
	scheduleFlightDataUpdates(tailNumbers, 5*time.Minute, 60*time.Second)

	// Set up HTTP server with endpoints
	http.HandleFunc("/flight", serveFlightInfo)
	http.Handle("/metrics", promhttp.Handler())

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := fmt.Sprintf(":%s", port)

	log.Printf("Starting flight tracker server on %s\n", serverAddr)
	log.Printf("Metrics available at /metrics\n")
	log.Printf("Flight info available at /flight?tail=<tail_number>\n")
	log.Printf("Tracking tail numbers: %s\n", strings.Join(tailNumbers, ", "))

	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
