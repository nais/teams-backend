package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

func Strp(s string) *string {
	return &s
}

func DerefString(s *string) string {
	if s == nil {
		return ""

	}
	return *s
}

func StringWithFallback(strp *string, fallback string) string {
	if strp == nil || *strp == "" {
		return fallback
	}
	return *strp
}

func DomainUsers(users []*sqlc.User, domain string) []*sqlc.User {
	domainUsers := make([]*sqlc.User, 0)
	suffix := "@" + domain

	for _, user := range users {
		if strings.HasSuffix(user.Email, suffix) {
			domainUsers = append(domainUsers, user)
		}
	}

	return domainUsers
}

// ServiceAccountEmail Generate a service account email address given the name of the service account
func ServiceAccountEmail(name slug.Slug, tenantDomain string) string {
	return string(name) + serviceAccountSuffix(tenantDomain)
}

// IsServiceAccount Check if a user is a service account or not
func IsServiceAccount(user dbmodels.User, tenantDomain string) bool {
	return strings.HasSuffix(user.Email, serviceAccountSuffix(tenantDomain))
}

// TeamPurpose Get a default team purpose
func TeamPurpose(purpose *string) string {
	return StringWithFallback(
		purpose,
		fmt.Sprintf("auto-generated by nais console on %s", time.Now().Format(time.RFC1123Z)),
	)
}

func serviceAccountSuffix(tenantDomain string) string {
	const serviceAccountSuffix = "serviceaccounts.nais.io"
	return "@" + tenantDomain + "." + serviceAccountSuffix
}
