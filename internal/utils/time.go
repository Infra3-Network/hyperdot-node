package utils

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidTimeRange = errors.New("invalid time range")
)

// Parse time range, time format
// 7h return 7 time.Hour
// 7d return 7 time.Hour * 24
func ParseTimeRange(tr string) (int, time.Duration, error) {
	var (
		amount int
		unit   time.Duration
	)

	if len(tr) < 2 {
		return 0, 0, ErrInvalidTimeRange
	}

	switch tr[len(tr)-1] {
	case 'h':
		unit = time.Hour
	case 'd':
		unit = time.Hour * 24
	default:
		return 0, 0, ErrInvalidTimeRange
	}

	_, err := fmt.Sscanf(tr, "%d", &amount)
	if err != nil {
		return 0, 0, err
	}

	return amount, unit, nil
}

// Get the time n hours ago
func GetTimeBefore(n int) time.Time {
	offset := time.Hour * time.Duration(n)
	return time.Now().Add(-offset)
}

// Get the time n days ago
func GetTimeBeforeDays(d int) time.Time {
	offset := time.Hour * 24 * time.Duration(d)
	return time.Now().Add(-offset)
}

// FormatTimeRange parse tr and return formated time
func FormatTimeRange(tr string) (string, error) {
	tn, td, err := ParseTimeRange(tr)
	if err != nil {
		return "", err
	}

	format := "2006-01-02 15:04:05"

	if td == time.Hour {
		return GetTimeBefore(tn).Format(format), nil
	} else if td == time.Hour*24 {
		return GetTimeBeforeDays(tn).Format(format), nil
	} else {
		return "", ErrInvalidTimeRange
	}
}
