package google_gcp_reconciler

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// ProjectId: Immutable. The unique, user-assigned id of the project. It
// must be 6 to 30 lowercase ASCII letters, digits, or hyphens. It must
// start with a letter. Trailing hyphens are prohibited. Example:
// `tokyo-rain-123`
func CreateProjectID(domain, environment, teamname string) string {
	hasher := sha256.New()
	hasher.Write([]byte(teamname))
	hasher.Write([]byte{0})
	hasher.Write([]byte(environment))
	hasher.Write([]byte{0})
	hasher.Write([]byte(domain))

	parts := make([]string, 3)
	parts[0] = truncate(teamname, 20)
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
