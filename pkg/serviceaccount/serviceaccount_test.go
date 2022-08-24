package serviceaccount_test

import (
	"testing"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/serviceaccount"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
)

func TestServiceAccountEmail(t *testing.T) {
	assert.Equal(t, "foo@domain.serviceaccounts.nais.io", serviceaccount.Email("foo", "domain"))
}

func TestIsServiceAccount(t *testing.T) {
	domainUser := db.User{
		User: &sqlc.User{
			Email: "user@domain.serviceaccounts.nais.io",
		},
	}
	exampleComUser := db.User{
		User: &sqlc.User{
			Email: "user@example.com.serviceaccounts.nais.io",
		},
	}

	assert.True(t, serviceaccount.IsServiceAccount(domainUser, "domain"))
	assert.False(t, serviceaccount.IsServiceAccount(exampleComUser, "domain"))
	assert.False(t, serviceaccount.IsServiceAccount(domainUser, "example.com"))
	assert.True(t, serviceaccount.IsServiceAccount(exampleComUser, "example.com"))
}
