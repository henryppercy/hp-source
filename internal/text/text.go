package text

import "strings"

func Slug(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return ' '
	}, s)
	return strings.Join(strings.Fields(s), "-")
}

// WordCount is the number of whitespace-separated words in s.
func WordCount(s string) int {
	return len(strings.Fields(s))
}

// ReadMinutes estimates reading time at 200 words a minute, rounded up to a
// whole minute and never below one.
func ReadMinutes(s string) int {
	mins := (WordCount(s) + 199) / 200
	if mins < 1 {
		return 1
	}
	return mins
}
