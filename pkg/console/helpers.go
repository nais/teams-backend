package console

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/nais/console/pkg/slug"
)

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

func SlugHashPrefixTruncate(slug slug.Slug, prefix string, maxLength int) string {
	hasher := sha256.New()
	hasher.Write([]byte(slug))

  prefixLength :=len(prefix)
  hashLength := 4
  slugLength := maxLength - prefixLength - hashLength - 2 // 2 becasue we join parts with '-'

	parts := []string{
		prefix,
		strings.TrimSuffix(Truncate(string(slug), slugLength), "-"),
		Truncate(hex.EncodeToString(hasher.Sum(nil)), hashLength),
	}

  return strings.Join(parts, "-")
}
