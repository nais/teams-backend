package serviceaccount

import (
	"strings"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/slug"
)

// Email Generate a service account email address given the name of the service account
func Email(name slug.Slug, tenantDomain string) string {
	return string(name) + serviceAccountSuffix(tenantDomain)
}

// IsServiceAccount Check if a user is a service account or not
func IsServiceAccount(user db.User, tenantDomain string) bool {
	return strings.HasSuffix(user.Email, serviceAccountSuffix(tenantDomain))
}

func serviceAccountSuffix(tenantDomain string) string {
	const serviceAccountSuffix = "serviceaccounts.nais.io"
	return "@" + tenantDomain + "." + serviceAccountSuffix
}
