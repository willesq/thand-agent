package common

import (
	"fmt"
	"strings"
	"time"

	iso8601arse "github.com/senseyeio/duration"
)

func ValidateDuration(duration string) (time.Duration, error) {
	w, err := validateDuration(duration)
	if err != nil {
		return 0, err
	}
	if w < 1*time.Minute {
		return 0, fmt.Errorf("duration must be at least 1 minutes")
	}
	return w, nil
}

func validateDuration(duration string) (time.Duration, error) {

	duration = strings.TrimSpace(duration)

	if parsedDuration, err := time.ParseDuration(duration); err == nil {
		return parsedDuration, nil
	} else if isoDuration, err := iso8601arse.ParseISO8601(duration); err == nil {
		referenceTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		shiftedTime := isoDuration.Shift(referenceTime)
		return shiftedTime.Sub(referenceTime), nil
	}

	return 0, fmt.Errorf("invalid duration format: %s. Expect ISO 8601 or duration string", duration)
}

// FormatDuration formats a duration in ISO 8601 format (PT1H30M20S)
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "PT0S"
	}

	var result strings.Builder
	result.WriteString("PT")

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		result.WriteString(fmt.Sprintf("%dH", hours))
	}
	if minutes > 0 {
		result.WriteString(fmt.Sprintf("%dM", minutes))
	}
	if seconds > 0 || (hours == 0 && minutes == 0) {
		result.WriteString(fmt.Sprintf("%dS", seconds))
	}

	return result.String()
}

// FormatHumanDuration formats a duration in human readable format (1 day, 2 hours, 3 minutes, 4 seconds)
func FormatDurationRemaining(d time.Duration) string {
	if d == 0 {
		return "0 seconds"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	var parts []string

	if days > 0 {
		if days == 1 {
			parts = append(parts, "1 day")
		} else {
			parts = append(parts, fmt.Sprintf("%d days", days))
		}
	}

	if hours > 0 {
		if hours == 1 {
			parts = append(parts, "1 hour")
		} else {
			parts = append(parts, fmt.Sprintf("%d hours", hours))
		}
	}

	if minutes > 0 {
		if minutes == 1 {
			parts = append(parts, "1 minute")
		} else {
			parts = append(parts, fmt.Sprintf("%d minutes", minutes))
		}
	}

	if seconds > 0 {
		if seconds == 1 {
			parts = append(parts, "1 second")
		} else {
			parts = append(parts, fmt.Sprintf("%d seconds", seconds))
		}
	}

	return strings.Join(parts, ", ")
}
