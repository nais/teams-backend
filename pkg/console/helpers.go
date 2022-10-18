package console

func Strp(s string) *string {
	return &s
}

func StringWithFallback(strp *string, fallback string) string {
	if strp == nil || *strp == "" {
		return fallback
	}
	return *strp
}

// Truncate will truncate the string s to the given length
func Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}
