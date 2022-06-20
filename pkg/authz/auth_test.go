package authz_test

import (
	"context"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContextWithUser(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, authz.UserFromContext(ctx))

	user := &dbmodels.User{
		Email: "mail@example.com",
		Name:  "User Name",
	}

	ctx = authz.ContextWithUser(ctx, user)
	assert.Equal(t, user, authz.UserFromContext(ctx))
}
