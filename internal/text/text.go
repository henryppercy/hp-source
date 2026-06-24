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
