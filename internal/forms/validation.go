package forms

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func validateDateOrYear(s string) error {
	if len(s) == 4 {
		if _, err := strconv.Atoi(s); err == nil {
			return nil
		}
	}

	if _, err := time.Parse("2006-01-02", s); err == nil {
		return nil
	}

	return fmt.Errorf("must be a year (e.g. 1963) or date (e.g. 1963-11-22)")
}

func validateDate(s string) error {
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return nil
	}

	return fmt.Errorf("must be a date (e.g. 1963-11-22)")
}

func validateDateOptional(s string) error {
	if s == "" {
		return nil
	}
	return validateDate(s)
}

func validateDateTimeOptional(s string) error {
	if s == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04", "2006-01-02 15:04:05"} {
		if _, err := time.Parse(layout, s); err == nil {
			return nil
		}
	}
	return fmt.Errorf("must be a date or date time, e.g. 2026-06-20 or 2026-06-20 14:32")
}

func validateCountryCode(s string) error {
	if s == "" {
		return nil
	}
	if len(strings.TrimSpace(s)) != 2 {
		return fmt.Errorf("must be a 2-letter code, e.g. GB")
	}
	return nil
}

func validateOptionalFloat(s string) error {
	if s == "" {
		return nil
	}
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return fmt.Errorf("must be a number, e.g. 53.22")
	}
	return nil
}

func validateOptionalPage(s string) error {
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return fmt.Errorf("must be a positive page number")
	}
	return nil
}
