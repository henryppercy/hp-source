package repo

import "testing"

func TestRatingDisplay(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{10, "5"},
		{8, "4"},
		{1, "0.5"},
		{0, ""},
		{11, ""},
	}
	for _, c := range cases {
		if got := RatingDisplay(c.in); got != c.want {
			t.Errorf("RatingDisplay(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRatingOptionsRoundTrip(t *testing.T) {
	opts := RatingOptions()
	if len(opts) == 0 {
		t.Fatal("RatingOptions() is empty")
	}
	for _, o := range opts {
		if got := RatingDisplay(o.Value); got != o.Label {
			t.Errorf("RatingDisplay(%d) = %q, want option label %q", o.Value, got, o.Label)
		}
	}
}
