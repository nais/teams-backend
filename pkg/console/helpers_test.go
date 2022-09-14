package console_test

import (
	"testing"

	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestStringWithFallback(t *testing.T) {
	t.Run("Fallback not used", func(t *testing.T) {
		assert.Equal(t, "some value", helpers.StringWithFallback(helpers.Strp("some value"), "some fallback value"))
	})

	t.Run("Fallback used", func(t *testing.T) {
		assert.Equal(t, "some fallback value", helpers.StringWithFallback(helpers.Strp(""), "some fallback value"))
	})
}

func TestStrp(t *testing.T) {
	s := "some string"
	assert.Equal(t, &s, helpers.Strp(s))
}

func TestDomainUsers(t *testing.T) {
	t.Run("No users", func(t *testing.T) {
		users := make([]*db.User, 0)
		domainUsers := helpers.DomainUsers(users, "example.com")
		assert.Len(t, domainUsers, 0)
	})

	t.Run("No users removed", func(t *testing.T) {
		users := []*db.User{
			{
				Email: "user1@example.com",
			},
			{
				Email: "user2@example.com",
			},
			{
				Email: "user3@example.com",
			},
		}
		domainUsers := helpers.DomainUsers(users, "example.com")
		assert.Len(t, domainUsers, 3)
		assert.Equal(t, "user1@example.com", domainUsers[0].Email)
		assert.Equal(t, "user2@example.com", domainUsers[1].Email)
		assert.Equal(t, "user3@example.com", domainUsers[2].Email)
	})

	t.Run("Users removed", func(t *testing.T) {
		users := []*db.User{
			{
				Email: "user1@example.com",
			},
			{
				Email: "user2@foo.bar",
			},
			{
				Email: "user3@example.com",
			},
		}
		domainUsers := helpers.DomainUsers(users, "example.com")
		assert.Len(t, domainUsers, 2)
		assert.Equal(t, "user1@example.com", domainUsers[0].Email)
		assert.Equal(t, "user3@example.com", domainUsers[1].Email)

		domainUsers = helpers.DomainUsers(users, "foo.bar")
		assert.Len(t, domainUsers, 1)
		assert.Equal(t, "user2@foo.bar", domainUsers[0].Email)
	})

	t.Run("User with missing email", func(t *testing.T) {
		users := []*db.User{
			{
				Name: "some name",
			},
			{
				Email: "user1@example.com",
			},
		}
		domainUsers := helpers.DomainUsers(users, "example.com")
		assert.Len(t, domainUsers, 1)
		assert.Equal(t, "user1@example.com", domainUsers[0].Email)
	})
}

func TestTruncate(t *testing.T) {
	t.Run("Empty string", func(t *testing.T) {
		assert.Equal(t, "", helpers.Truncate("", 5))
	})

	t.Run("String shorter than truncate length", func(t *testing.T) {
		assert.Equal(t, "some string", helpers.Truncate("some string", 20))
	})

	t.Run("String longer than truncate length", func(t *testing.T) {
		assert.Equal(t, "some ", helpers.Truncate("some string", 5))
	})
}
