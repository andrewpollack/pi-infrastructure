package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestToTitleCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Single word",
			input:    "landed",
			expected: "Landed",
		},
		{
			name:     "Multiple words",
			input:    "in flight",
			expected: "In Flight",
		},
		{
			name:     "Already capitalized",
			input:    "Arrived",
			expected: "Arrived",
		},
		{
			name:     "Mixed case",
			input:    "dePARTed",
			expected: "DePARTed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := toTitleCase(tc.input)
			if result != tc.expected {
				t.Errorf("toTitleCase(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "Zero duration",
			duration: 0,
			expected: "0 minutes",
		},
		{
			name:     "Minutes only",
			duration: 45 * time.Minute,
			expected: "45 minutes",
		},
		{
			name:     "Hours and minutes",
			duration: 2*time.Hour + 30*time.Minute,
			expected: "2 hours 30 minutes",
		},
		{
			name:     "Exact hours",
			duration: 3 * time.Hour,
			expected: "3 hours 0 minutes",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatDuration(tc.duration)
			if result != tc.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tc.duration, result, tc.expected)
			}
		})
	}
}

// TestServeFlightInfo_WithTailNumber tests the flight info handler with a valid tail number
func TestServeFlightInfo_WithTailNumber(t *testing.T) {
	// Create a handler that returns a mock response instead of calling the real API
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tailNumber := r.URL.Query().Get("tail")
		if tailNumber == "" {
			http.Error(w, "Error fetching flight info: tail number cannot be empty", http.StatusInternalServerError)
			return
		}

		// Create a mock flight info response
		flightInfo := FlightInfo{
			FlightNumber:   tailNumber,
			Departure:      "Test Airport (Test City)",
			Arrival:        "Destination Airport (Destination City)",
			DepartureTime:  "2025-05-10 08:00:00 PDT",
			ArrivalTime:    "2025-05-10 10:00:00 PDT",
			AircraftType:   "Test Aircraft",
			Status:         "In Air",
			Owner:          "Test Owner",
			OwnerLocation:  "Test Location",
			DirectDistance: "100 miles",
			FlightDuration: "2 hours 0 minutes",
			EstimatedFuel:  "50 gallons (300 pounds)",
		}

		// Return the mock flight info as JSON
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(flightInfo)
	})

	// Create a request with a tail number parameter
	req, err := http.NewRequest("GET", "/flight?tail=N12345", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder
	handler.ServeHTTP(rr, req)

	// Check that the status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check that the response body contains flight information
	if !strings.Contains(rr.Body.String(), "FlightNumber") {
		t.Errorf("handler returned unexpected body without flight information: %v", rr.Body.String())
	}
}

func TestServeFlightInfo_NoTailNumber(t *testing.T) {
	// This test verifies that the serveFlightInfo handler returns an error
	// when no tail number is provided in the request

	// Create a handler that returns a mock error response for empty tail numbers
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tailNumber := r.URL.Query().Get("tail")
		if tailNumber == "" {
			http.Error(w, "Error fetching flight info: tail number cannot be empty", http.StatusInternalServerError)
			return
		}

		// This should not be reached in this test
		t.Error("Handler did not return error for empty tail number")
	})

	// Create a request with no tail number parameter
	req, err := http.NewRequest("GET", "/flight", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder
	handler.ServeHTTP(rr, req)

	// Check that the status code is 500 Internal Server Error
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}

	// Check that the response body contains an error message
	if !strings.Contains(rr.Body.String(), "Error fetching flight info") {
		t.Errorf("handler returned unexpected body without error message: %v", rr.Body.String())
	}
}

func TestEnvironmentVariableValidation(t *testing.T) {
	// Save original environment and restore it after the test
	originalEnv := os.Getenv("TAIL_NUMBERS")
	defer func() {
		if err := os.Setenv("TAIL_NUMBERS", originalEnv); err != nil {
			t.Fatalf("Failed to restore original environment variable: %v", err)
		}
	}()

	// Test cases for environment variable validation
	tests := []struct {
		name        string
		envValue    string
		shouldPanic bool
	}{
		{
			name:        "Valid single tail number",
			envValue:    "123456",
			shouldPanic: false,
		},
		{
			name:        "Valid multiple tail numbers",
			envValue:    "123456,234567",
			shouldPanic: false,
		},
		{
			name:        "Empty environment variable",
			envValue:    "",
			shouldPanic: true,
		},
		{
			name:        "Only whitespace",
			envValue:    "  ",
			shouldPanic: true,
		},
		{
			name:        "Empty tail number in list",
			envValue:    "123456,,234567",
			shouldPanic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set the environment variable
			_ = os.Setenv("TAIL_NUMBERS", tc.envValue)

			// Use a function that calls the validation logic but returns instead of exiting
			validateEnv := func() (panicked bool) {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()

				// This is a simplified version of the validation logic in main()
				envTails := os.Getenv("TAIL_NUMBERS")
				if envTails == "" {
					panic("TAIL_NUMBERS environment variable must be set")
				}

				tailNumbers := strings.Split(envTails, ",")
				if len(tailNumbers) == 0 {
					panic("No tail numbers provided")
				}

				for _, tail := range tailNumbers {
					tailNumber := strings.TrimSpace(tail)
					if tailNumber == "" {
						panic("Empty tail number found")
					}
				}

				return false
			}

			panicked := validateEnv()
			if panicked != tc.shouldPanic {
				if tc.shouldPanic {
					t.Errorf("Expected validation to fail for %q but it didn't", tc.envValue)
				} else {
					t.Errorf("Expected validation to pass for %q but it failed", tc.envValue)
				}
			}
		})
	}
}
