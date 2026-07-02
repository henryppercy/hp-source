package repo

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableInt(n int) *int {
	if n == 0 {
		return nil
	}
	return &n
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
