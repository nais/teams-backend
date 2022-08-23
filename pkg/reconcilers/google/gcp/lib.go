package google_gcp_reconciler

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/nais/console/pkg/slug"
	"strings"
)

// GenerateProjectID Generate a unique project ID for the team in a given environment in a deterministic fashion
func GenerateProjectID(domain, environment string, slug slug.Slug) string {
	hasher := sha256.New()
	hasher.Write([]byte(slug))
	hasher.Write([]byte{0})
	hasher.Write([]byte(environment))
	hasher.Write([]byte{0})
	hasher.Write([]byte(domain))

	parts := make([]string, 3)
	parts[0] = truncate(string(slug), 20)
	parts[1] = truncate(environment, 4)
	parts[2] = truncate(hex.EncodeToString(hasher.Sum(nil)), 4)

	return strings.Join(parts, "-")
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}
