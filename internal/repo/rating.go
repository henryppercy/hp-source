package repo

// RatingOption pairs a stored rating value with its display label.
type RatingOption struct {
	Value int
	Label string
}

var ratingOptions = []RatingOption{
	{10, "5"}, {9, "4.5"}, {8, "4"}, {7, "3.5"}, {6, "3"},
	{5, "2.5"}, {4, "2"}, {3, "1.5"}, {2, "1"}, {1, "0.5"},
}

// RatingOptions returns the rating choices, highest first.
func RatingOptions() []RatingOption {
	return ratingOptions
}

// RatingDisplay returns the display label for a stored rating value, or ""
// for an unknown one.
func RatingDisplay(rating int) string {
	for _, o := range ratingOptions {
		if o.Value == rating {
			return o.Label
		}
	}
	return ""
}
