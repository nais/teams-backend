package usersync_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/test"
	"github.com/nais/console/pkg/usersync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSync(t *testing.T) {
	domain := "example.com"

	t.Run("Server error from Google", func(t *testing.T) {
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)

		httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
			return test.Response("500 Internal Server Error", `{"error": "some error"}`)
		})

		usersync := usersync.New(database, auditLogger, domain, httpClient)
		err := usersync.Sync(context.Background())
		assert.ErrorContains(t, err, "list remote users")
	})

	t.Run("No remote users", func(t *testing.T) {
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)

		database.
			On("Transaction", mock.Anything, mock.Anything).
			Return(nil).
			Once()

		httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
			return test.Response("200 OK", `{"users":[]}`)
		})

		usersync := usersync.New(database, auditLogger, domain, httpClient)
		err := usersync.Sync(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Create, update and delete users", func(t *testing.T) {
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		dbtx := db.NewMockDatabase(t)

		numDefaultRoleNames := len(usersync.DefaultRoleNames)

		localUserWithIncorrectName := getUser("user1@example.com", "Outdated Name", nil)
		updatedLocalUser := getUser(localUserWithIncorrectName.Email, "Correct Name", &localUserWithIncorrectName.ID)
		newLocalUser := getUser("user2@example.com", "Create Me", nil)
		localUserThatWillBeDeleted := getUser("delete-me@example.com", "Delete Me", nil)

		httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
			return test.Response("200 OK", `{"users":[`+
				`{"primaryEmail":"user1@example.com","name":{"fullName":"Correct Name"}},`+
				`{"primaryEmail":"user2@example.com","name":{"fullName":"Create Me"}}]}`) // Will be created
		})

		ctx := context.Background()
		txCtx := context.Background()

		database.
			On("Transaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(db.TransactionFunc)
				fn(txCtx, dbtx)
			}).
			Return(nil).
			Once()

		dbtx.
			On("GetUserByEmail", txCtx, localUserWithIncorrectName.Email).
			Return(localUserWithIncorrectName, nil).
			Once()

		dbtx.
			On("SetUserName", txCtx, localUserWithIncorrectName.ID, "Correct Name").
			Return(updatedLocalUser, nil).
			Once()

		dbtx.
			On("AssignGlobalRoleToUser", txCtx, updatedLocalUser.ID, mock.AnythingOfType("sqlc.RoleName")).
			Return(nil).
			Times(numDefaultRoleNames)

		dbtx.
			On("GetUserByEmail", txCtx, newLocalUser.Email).
			Return(nil, errors.New("user not found")).
			Once()

		dbtx.
			On("AddUser", txCtx, newLocalUser.Name, newLocalUser.Email).
			Return(newLocalUser, nil).
			Once()

		dbtx.
			On("AssignGlobalRoleToUser", txCtx, newLocalUser.ID, mock.AnythingOfType("sqlc.RoleName")).
			Return(nil).
			Times(numDefaultRoleNames)

		dbtx.
			On("GetUsersByEmail", txCtx, "%@example.com").
			Return([]*db.User{
				updatedLocalUser, newLocalUser, localUserThatWillBeDeleted,
			}, nil).
			Once()

		dbtx.
			On("DeleteUser", txCtx, localUserThatWillBeDeleted.ID).
			Return(nil).
			Once()

		auditLogger.
			On("Logf", ctx, mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionUsersyncUpdate &&
					*f.TargetUserEmail == updatedLocalUser.Email
			}), "Local user updated: user1@example.com").
			Return(nil).
			Once()

		auditLogger.
			On("Logf", ctx, mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionUsersyncCreate &&
					*f.TargetUserEmail == newLocalUser.Email
			}), "Local user created: user2@example.com").
			Return(nil).
			Once()

		auditLogger.
			On("Logf", ctx, mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionUsersyncDelete &&
					*f.TargetUserEmail == localUserThatWillBeDeleted.Email
			}), "Local user deleted: delete-me@example.com").
			Return(nil).
			Once()

		usersync := usersync.New(database, auditLogger, domain, httpClient)
		err := usersync.Sync(ctx)
		assert.NoError(t, err)
	})
}

func getUser(email, name string, id *uuid.UUID) *db.User {
	if id == nil {
		newID := uuid.New()
		id = &newID
	}
	return &db.User{
		ID:    *id,
		Email: email,
		Name:  name,
	}
}
