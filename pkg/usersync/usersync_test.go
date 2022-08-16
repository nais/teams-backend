package usersync_test

import (
	"context"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/roles"
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

	t.Run("Handle conflict for user roles", func(t *testing.T) {
		db, _ := test.GetTestDBWithRoles()
		db.Create(system)
		user := &dbmodels.User{Email: "user@example.com"}
		role := &dbmodels.Role{}
		db.Where("name = ?", roles.RoleTeamCreator).First(role)
		db.Create([]*dbmodels.User{user})
		db.Create(&dbmodels.UserRole{RoleID: *role.ID, UserID: *user.ID})

		httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
			return test.Response("200 OK", `{"users":[`+
				`{"primaryEmail":"user@example.com","name":{"fullName":"User Name"}}]}`) // Will be created
		})

		mockAuditLogger.
			On("Logf", usersync.OpUpdate, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(targetUser *dbmodels.User) bool {
				return targetUser.Name == "User Name"
			}), mock.Anything).
			Return(nil).
			Once()

		usersync := usersync.New(db, *system, mockAuditLogger, "example.com", httpClient)
		err := usersync.Sync(context.Background())
		assert.NoError(t, err)
		mockAuditLogger.AssertExpectations(t)
	})
}
