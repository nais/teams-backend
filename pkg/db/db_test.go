package db

import (
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUserByEmail(t *testing.T) {
	db := test.GetTestDB()
	db.AutoMigrate(&dbmodels.User{})
	user := &dbmodels.User{
		Name:  "User",
		Email: "user@example.com",
	}
	db.Create(user)

	assert.Nil(t, GetUserByEmail(db, "user-that-does-not-exist@example.com"))
	assert.Equal(t, user.ID, GetUserByEmail(db, "User@example.com").ID)
	assert.Equal(t, user.ID, GetUserByEmail(db, "user@example.com").ID)
}
