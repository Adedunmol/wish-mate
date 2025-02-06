package helpers_test

import (
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"math"
	"testing"
	"time"
)

func TestCalculateDaysBefore(t *testing.T) {

	t.Run("calculate the exact date before a given date", func(t *testing.T) {

		got, _ := helpers.CalculateDaysBefore("2025-03-06", 2)
		date := time.Date(time.Now().Year(), 3, 6, 0, 0, 0, 0, time.UTC)

		if got.Year() > time.Now().Year() {
			diff := math.Abs(float64(got.Year() - time.Now().Year()))
			date = date.AddDate(int(diff), 0, 0)
		}

		want := time.Date(date.Year(), date.Month(), date.AddDate(0, 0, -2).Day(), 0, 0, 0, 0, time.UTC)
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("calculate the exact date before a given date for beginning of month", func(t *testing.T) {

		got, _ := helpers.CalculateDaysBefore("2025-03-01", 2)

		date := time.Date(time.Now().Year(), 3, 1, 0, 0, 0, 0, time.UTC)

		if got.Year() > time.Now().Year() {
			diff := math.Abs(float64(got.Year() - time.Now().Year()))
			date = date.AddDate(int(diff), 0, 0)
		}

		want := time.Date(date.Year(), date.AddDate(0, -1, 0).Month(), date.AddDate(0, 0, -2).Day(), 0, 0, 0, 0, time.UTC)
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("calculate the exact date (next year) before a given date for past birthday", func(t *testing.T) {

		got, _ := helpers.CalculateDaysBefore("2005-02-06", 2)

		date := time.Date(time.Now().Year(), 2, 6, 0, 0, 0, 0, time.UTC)

		if got.Year() > time.Now().Year() {
			diff := math.Abs(float64(got.Year() - time.Now().Year()))
			date = date.AddDate(int(diff), 0, 0)
		}

		want := time.Date(date.Year(), date.Month(), date.AddDate(0, 0, -2).Day(), 0, 0, 0, 0, time.UTC)
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("return error for invalid month", func(t *testing.T) {
		_, err := helpers.CalculateDaysBefore("2025-14-10", 2)

		if err == nil {
			t.Errorf("expected error for month out of range")
		}
	})

	t.Run("return error for invalid day", func(t *testing.T) {
		_, err := helpers.CalculateDaysBefore("2025-02-30", 2)

		if err == nil {
			t.Errorf("expected error for day out of range")
		}
	})

	t.Run("return error for invalid format", func(t *testing.T) {
		_, err := helpers.CalculateDaysBefore("2025/02/30", 2)

		if err == nil {
			t.Errorf("expected error for invalid format")
		}
	})

	t.Run("return error for negative notifyBefore", func(t *testing.T) {
		_, err := helpers.CalculateDaysBefore("2025-02-05", -1)

		if err == nil {
			t.Errorf("expected error for negative notifyBefore")
		}
	})
}
