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
