package usersync_test

import (
	"context"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/test"
	"github.com/nais/console/pkg/usersync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

func TestSync(t *testing.T) {
	system := &dbmodels.System{Name: "console"}
	mockAuditLogger := &auditlogger.MockAuditLogger{}

	t.Run("Server error from Google", func(t *testing.T) {
		db, _ := test.GetTestDB()
		db.Create(system)

		httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
			return test.Response("500 Internal Server Error", `{"error": "some error"}`)
		})

		usersync := usersync.New(db, *system, mockAuditLogger, "example.com", httpClient)
		err := usersync.Sync(context.Background())
		assert.ErrorContains(t, err, "usersync:list:remote: list remote users")
	})

	t.Run("No remote users", func(t *testing.T) {
		db, _ := test.GetTestDB()
		db.Create(system)

		httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
			return test.Response("200 OK", `{"users":[]}`)
		})

		usersync := usersync.New(db, *system, mockAuditLogger, "example.com", httpClient)
		err := usersync.Sync(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Create, update and delete users", func(t *testing.T) {
		db, _ := test.GetTestDB()
		db.Create(system)
		db.Create([]*dbmodels.User{
			{Email: "delete-me@example.com", Name: "Delete Me"},   // Will be deleted
			{Email: "dont-delete-me@service-account.example.com"}, // Will not be touched
			{Email: "user1@example.com", Name: "Update Me"},       // Will be updated
		})

		httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
			return test.Response("200 OK", `{"users":[`+
				`{"primaryEmail":"user1@example.com","name":{"fullName":"Correct Name"}},`+
				`{"primaryEmail":"user2@example.com","name":{"fullName":"Create Me"}}]}`) // Will be created
		})

		mockAuditLogger.
			On("Logf", usersync.OpUpdate, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(targetUser *dbmodels.User) bool {
				return targetUser.Name == "Correct Name"
			}), mock.Anything).
			Return(nil).
			Once()

		mockAuditLogger.
			On("Logf", usersync.OpCreate, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(targetUser *dbmodels.User) bool {
				return targetUser.Name == "Create Me"
			}), mock.Anything).
			Return(nil).
			Once()

		mockAuditLogger.
			On("Logf", usersync.OpDelete, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(targetUser *dbmodels.User) bool {
				return targetUser.Name == "Delete Me"
			}), mock.Anything).
			Return(nil).
			Once()

		usersync := usersync.New(db, *system, mockAuditLogger, "example.com", httpClient)
		err := usersync.Sync(context.Background())
		assert.NoError(t, err)
		mockAuditLogger.AssertExpectations(t)
	})

	t.Run("Don't insert duplicate role bindings", func(t *testing.T) {
		db, _ := test.GetTestDBWithRoles()
		db.Create(system)
		user1 := &dbmodels.User{Email: "user1@example.com"}
		user2 := &dbmodels.User{Email: "user2@example.com"}
		db.Create([]*dbmodels.User{user1, user2})

		roundtripFunc := func(req *http.Request) *http.Response {
			return test.Response("200 OK", `{"users":[`+
				`{"primaryEmail":"user1@example.com","name":{"fullName":"Name 1"}},`+
				`{"primaryEmail":"user2@example.com","name":{"fullName":"Name 2"}}`+
				`]}`)
		}
		httpClient := test.NewTestHttpClient(roundtripFunc, roundtripFunc)

		ctx := context.Background()
		us := usersync.New(db, *system, auditlogger.New(db), "example.com", httpClient)
		assert.NoError(t, us.Sync(ctx))
		assert.NoError(t, us.Sync(ctx)) // Run twice, should only add role bindngs once

		var count1, count2 int64
		db.Model(&dbmodels.UserRole{}).Where("user_id = ?", user1.ID).Count(&count1)
		db.Model(&dbmodels.UserRole{}).Where("user_id = ?", user2.ID).Count(&count2)

		assert.EqualValues(t, len(usersync.DefaultRoleNames), count1)
		assert.EqualValues(t, len(usersync.DefaultRoleNames), count2)
	})
}
