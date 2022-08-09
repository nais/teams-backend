package db_test

import (
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUserByEmail(t *testing.T) {
	testDb, _ := test.GetTestDB()
	user := &dbmodels.User{
		Name:  "User",
		Email: "user@example.com",
	}
	testDb.Create(user)

	assert.Nil(t, db.GetUserByEmail(testDb, "user-that-does-not-exist@example.com"))
	assert.Equal(t, user.ID, db.GetUserByEmail(testDb, "User@example.com").ID)
	assert.Equal(t, user.ID, db.GetUserByEmail(testDb, "user@example.com").ID)
}
