package console

import (
	"github.com/nais/console/pkg/dbmodels"
	"strings"
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

func DomainUsers(users []*dbmodels.User, domain string) []*dbmodels.User {
	domainUsers := make([]*dbmodels.User, 0)
	suffix := "@" + domain

	for _, user := range users {
		if user.Email != nil && strings.HasSuffix(*user.Email, suffix) {
			domainUsers = append(domainUsers, user)
		}
	}

	return domainUsers
}

// ServiceAccountEmail Generate a service account email address given the name of the service account
func ServiceAccountEmail(name dbmodels.Slug, partnerDomain string) string {
	return string(name) + serviceAccountSuffix(partnerDomain)
}

// IsServiceAccount Check if a user is a service account or not
func IsServiceAccount(user dbmodels.User, partnerDomain string) bool {
	return strings.HasSuffix(DerefString(user.Email), serviceAccountSuffix(partnerDomain))
}

func serviceAccountSuffix(partnerDomain string) string {
	const serviceAccountSuffix = "serviceaccounts.nais.io"
	return "@" + partnerDomain + "." + serviceAccountSuffix
}
