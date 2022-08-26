package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/nais/console/pkg/db"
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

// DomainUsers Return users in a list of of user object that has an email address with the tenant domain as suffix
func DomainUsers(users []*db.User, domain string) []*db.User {
	domainUsers := make([]*db.User, 0)
	suffix := "@" + domain

	for _, user := range users {
		if strings.HasSuffix(user.Email, suffix) {
			domainUsers = append(domainUsers, user)
		}
	}

	return domainUsers
}

// TeamPurpose Get a default team purpose
func TeamPurpose(purpose *string) string {
	return StringWithFallback(
		purpose,
		fmt.Sprintf("auto-generated by nais console on %s", time.Now().Format(time.RFC1123Z)),
	)
}
