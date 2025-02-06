package helpers_test

import (
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"testing"
	"time"
)

func TestCalculateDaysBefore(t *testing.T) {

	t.Run("calculate the exact date before a given date", func(t *testing.T) {

		got, _ := helpers.CalculateDaysBefore("2025-02-06", 2)

		dayDuration := 24 * time.Hour
		want := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Add(-2*dayDuration).Day(), 0, 0, 0, 0, time.Local)
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("calculate the exact date (next year) before a given date for past birthday", func(t *testing.T) {

		got, _ := helpers.CalculateDaysBefore("2025-02-06", 2)

		dayDuration := 24 * time.Hour
		want := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Add(-2*dayDuration).Day(), 0, 0, 0, 0, time.Local)
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("return error for invalid date", func(t *testing.T) {})
}
