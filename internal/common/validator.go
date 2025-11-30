package common

import (
	"net/mail"
	"net/url"
	"strconv"
	"strings"
)

func IsValidLoginServer(hostname string) bool {

	// paarse url
	_, err := url.Parse(hostname)

	return err == nil
}

// IsAllDigits checks if a string contains only digits (0-9)
// This is optimized for speed by checking each byte directly
func IsAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

func IsValidURL(rawurl string) bool {
	_, err := url.ParseRequestURI(rawurl)
	return err == nil
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsValidNumber(value string, decimalAllowed bool) bool {
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}
	if !decimalAllowed && strings.Contains(value, ".") {
		return false
	}
	return true
}

// ValidateNumberRange checks if a number value is within the specified min/max range.
// Returns an error message if validation fails, or an empty string if validation passes.
// Empty minValue or maxValue strings are treated as no constraint.
func ValidateNumberRange(value string, minValue string, maxValue string) string {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "Please enter a valid number"
	}

	if len(minValue) != 0 {
		min, err := strconv.ParseFloat(minValue, 64)
		if err == nil && num < min {
			return "Value must be at least " + minValue
		}
	}

	if len(maxValue) != 0 {
		max, err := strconv.ParseFloat(maxValue, 64)
		if err == nil && num > max {
			return "Value must be at most " + maxValue
		}
	}

	return ""
}
