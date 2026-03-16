package forms

import (
	"fmt"
	"strconv"
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
