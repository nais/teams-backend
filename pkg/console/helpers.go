package helpers

func Strp(s string) *string {
	return &s
}

func StringWithFallback(strp *string, fallback string) string {
	if strp == nil || *strp == "" {
		return fallback
	}
	return *strp
}
