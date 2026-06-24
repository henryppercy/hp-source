package text

import "testing"

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
