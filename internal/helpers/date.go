package helpers

import (
	"fmt"
	"time"
)

func CalculateDaysBefore(dateOfBirth string, notifyBefore int) (time.Time, error) {

	dob, err := time.Parse("2006-01-02", dateOfBirth)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing date: use YYYY-MM-DD: %v", err)
	}

	now := time.Now()
	birthday := time.Date(now.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, time.UTC)

	var nextBirthday time.Time
	if birthday.Before(now) {
		// If this year's birthday has passed, calculate for the next year
		nextBirthday = birthday.AddDate(1, 0, 0)
	} else {
		// Otherwise, use this year's birthday
		nextBirthday = birthday
	}

	scheduledDate := nextBirthday.AddDate(0, 0, -notifyBefore)

	return scheduledDate, nil
}
