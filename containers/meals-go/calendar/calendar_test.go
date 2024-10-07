package calendar

import (
	"testing"
	"time"
)

func TestDaysInMonth(t *testing.T) {
	calendar := NewCalendar(2024, time.February)
	if days := calendar.DaysInMonth(); days != 29 {
		t.Errorf("Expected February 2024 to have 29 days, got %d", days)
	}
}

func TestFirstWeekdayOfMonth(t *testing.T) {
	calendar := NewCalendar(2024, time.October)
	if weekday := calendar.FirstWeekdayOfMonth(); weekday != time.Tuesday {
		t.Errorf("Expected October 2024 to start on a Tuesday, got %v", weekday)
	}

	calendar = NewCalendar(2024, time.November)
	if weekday := calendar.FirstWeekdayOfMonth(); weekday != time.Friday {
		t.Errorf("Expected November 2024 to start on a Friday, got %v", weekday)
	}
}

func TestWeekIndex(t *testing.T) {
	calendar := NewCalendar(2024, time.October)

	type DayToExpectedIndex struct {
		Day           int
		ExpectedIndex int
	}

	tuples := []DayToExpectedIndex{
		{4, 0},
		{5, 0},
		{6, 1},
		{12, 1},
		{13, 2},
		{18, 2},
		{20, 3},
		{25, 3},
		{27, 4},
		{31, 4},
	}

	for _, tuple := range tuples {
		// Get the week index of the given day
		weekIndex := calendar.GetWeekIndexOfDay(tuple.Day)

		// Assert that the returned week index matches the expected value
		if weekIndex != tuple.ExpectedIndex {
			t.Errorf("For day %d, expected week index %d, but got %d", tuple.Day, tuple.ExpectedIndex, weekIndex)
		}
	}
}

func TestPrintMonthCalendar(t *testing.T) {
	calendar := NewCalendar(2024, time.October)
	calendar.PrintMonthCalendar() // Just to verify it prints properly, no need for test comparison here.
}
