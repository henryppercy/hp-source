package text

import (
	"strings"
	"testing"
)

func TestSlug(t *testing.T) {
	cases := []struct{ in, want string }{
		{"The Hobbit!", "the-hobbit"},
		{"  A  B ", "a-b"},
		{"café", "caf"}, // ASCII-only: non-[a-z0-9] dropped
		{"1984", "1984"},
		{"", ""},
	}
	for _, c := range cases {
		if got := Slug(c.in); got != c.want {
			t.Errorf("Slug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestReadMinutes(t *testing.T) {
	cases := []struct {
		words, want int
	}{
		{0, 1},   // never below one
		{1, 1},   // rounds up
		{200, 1}, // exactly one minute
		{201, 2}, // over the line
		{600, 3},
	}
	for _, c := range cases {
		body := strings.TrimSpace(strings.Repeat("word ", c.words))
		if got := ReadMinutes(body); got != c.want {
			t.Errorf("ReadMinutes(%d words) = %d, want %d", c.words, got, c.want)
		}
	}
}
