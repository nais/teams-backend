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

// Contains Check if a specific value is in a slice of strings
func Contains(strings []string, contains string) bool {
	for _, value := range strings {
		if value == contains {
			return true
		}
	}
	return false
}
