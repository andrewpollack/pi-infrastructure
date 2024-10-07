package calendar

import (
	"fmt"
	"time"
)

type Day struct {
	Number int
}

type Calendar struct {
	Year  int
	Month time.Month
	Weeks [][]Day
}

func NewCalendar(year int, month time.Month) *Calendar {
	calendar := &Calendar{
		Year:  year,
		Month: month,
	}
	calendar.buildMonthCalendar()
	return calendar
}

// buildMonthCalendar constructs the internal representation of the calendar as a list of weeks.
func (c *Calendar) buildMonthCalendar() {
	var weeks [][]Day

	// Get the first day of the month and the number of days in the month
	firstDay := c.FirstWeekdayOfMonth()
	daysInMonth := c.DaysInMonth()

	// We will fill up each week (list) starting from Sunday
	week := make([]Day, 7)
	day_number := 1

	// Fill leading zeros for the first week if the month doesn't start on Sunday
	for i := 0; i < int(firstDay); i++ {
		week[i] = Day{0}
	}

	// Fill the calendar with days of the month
	for day_number <= daysInMonth {
		week[int(firstDay)] = Day{day_number}
		firstDay++
		day_number++

		// If the week is complete (i.e., Sunday to Saturday), add it to weeks and start a new week
		if firstDay > 6 {
			weeks = append(weeks, week)
			week = make([]Day, 7) // Start a new week
			firstDay = 0
		}
	}

	// Add the last incomplete week, if any
	if firstDay != 0 {
		weeks = append(weeks, week)
	}

	// Set the internal representation of weeks
	c.Weeks = weeks
}

func (c *Calendar) DaysInMonth() int {
	firstOfNextMonth := time.Date(c.Year, c.Month+1, 1, 0, 0, 0, 0, time.UTC)
	lastOfMonth := firstOfNextMonth.AddDate(0, 0, -1)
	return lastOfMonth.Day()
}

func (c *Calendar) FirstWeekdayOfMonth() time.Weekday {
	firstOfMonth := time.Date(c.Year, c.Month, 1, 0, 0, 0, 0, time.UTC)
	return firstOfMonth.Weekday()
}

func (c *Calendar) GetWeekIndexOfDay(day int) int {
	firstOfMonth := time.Date(c.Year, c.Month, 1, 0, 0, 0, 0, time.UTC)

	// Find the weekday for the first day of the month
	firstDayOfWeek := firstOfMonth.Weekday()

	// Calculate how many days have passed since the first of the month
	daysFromStart := day - 1

	// Calculate which week the current day falls into
	weekIndex := (daysFromStart + int(firstDayOfWeek)) / 7

	return weekIndex
}

// IsFriday checks if the given day in the current calendar month falls on a Friday.
func (c *Calendar) GetWeekday(day int) time.Weekday {
	// Create a time.Date object for the given day
	date := time.Date(c.Year, c.Month, day, 0, 0, 0, 0, time.UTC)

	return date.Weekday()
}

// PrintMonthCalendar prints the calendar for the current month.
func (c *Calendar) PrintMonthCalendar() {
	// Print the header for days of the week
	fmt.Println("Su Mo Tu We Th Fr Sa")

	// Loop through the weeks and print each day
	for _, week := range c.Weeks {
		for _, day := range week {
			if day.Number == 0 {
				fmt.Print("   ")
			} else {
				fmt.Printf("%2d ", day.Number)
			}
		}
		fmt.Println()
	}
}
