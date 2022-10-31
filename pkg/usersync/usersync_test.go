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

		localUserWithIncorrectName := &db.User{User: &sqlc.User{ID: serialUuid(1), Email: "user1@example.com", ExternalID: "123", Name: "Incorrect Name"}}
		localUserWithCorrectName := &db.User{User: &sqlc.User{ID: serialUuid(1), Email: "user1@example.com", ExternalID: "123", Name: "Correct Name"}}

		localUserWithIncorrectEmail := &db.User{User: &sqlc.User{ID: serialUuid(2), Email: "user-123@example.com", Name: "Some Name"}}
		localUserWithCorrectEmail := &db.User{User: &sqlc.User{ID: serialUuid(2), Email: "user3@example.com", Name: "Some Name"}}

		localUserThatWillBeDeleted := &db.User{User: &sqlc.User{ID: serialUuid(3), Email: "delete-me@example.com", Name: "Delete Me"}}

		createdLocalUser := &db.User{User: &sqlc.User{ID: serialUuid(4), Email: "user2@example.com", ExternalID: "456", Name: "Create Me"}}

		httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
			return test.Response("200 OK", `{"users":[`+
				`{"Id": "123", "primaryEmail":"user1@example.com","name":{"fullName":"Correct Name"}},`+ // Will update name of local user
				`{"Id": "456", "primaryEmail":"user2@example.com","name":{"fullName":"Create Me"}},`+ // Will be created
				`{"Id": "789", "primaryEmail":"user3@example.com","name":{"fullName":"Some Name"}}]}`) // Will update email of local user
		})

		ctx := context.Background()
		txCtx := context.Background()

		database.
			On("Transaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(db.DatabaseTransactionFunc)
				fn(txCtx, dbtx)
			}).
			Return(nil).
			Once()

		// user1@example.com
		dbtx.
			On("GetUserByExternalID", txCtx, "123").
			Return(localUserWithIncorrectName, nil).
			Once()
		dbtx.
			On("UpdateUser", txCtx, localUserWithIncorrectName.ID, "Correct Name", "user1@example.com", "123").
			Return(localUserWithCorrectName, nil).
			Once()
		dbtx.
			On("AssignGlobalRoleToUser", txCtx, localUserWithCorrectName.ID, mock.AnythingOfType("sqlc.RoleName")).
			Return(nil).
			Times(numDefaultRoleNames)

		// user2@example.com
		dbtx.
			On("GetUserByExternalID", txCtx, "456").
			Return(nil, errors.New("user not found")).
			Once()
		dbtx.
			On("GetUserByEmail", txCtx, "user2@example.com").
			Return(nil, errors.New("user not found")).
			Once()
		dbtx.
			On("CreateUser", txCtx, "Create Me", "user2@example.com", "456").
			Return(createdLocalUser, nil).
			Once()
		dbtx.
			On("AssignGlobalRoleToUser", txCtx, createdLocalUser.ID, mock.AnythingOfType("sqlc.RoleName")).
			Return(nil).
			Times(numDefaultRoleNames)

		// user3@example.com
		dbtx.
			On("GetUserByExternalID", txCtx, "789").
			Return(localUserWithIncorrectEmail, nil).
			Once()
		dbtx.
			On("UpdateUser", txCtx, localUserWithIncorrectEmail.ID, "Some Name", "user3@example.com", "789").
			Return(localUserWithCorrectEmail, nil).
			Once()
		dbtx.
			On("AssignGlobalRoleToUser", txCtx, localUserWithCorrectEmail.ID, mock.AnythingOfType("sqlc.RoleName")).
			Return(nil).
			Times(numDefaultRoleNames)

		dbtx.
			On("GetUsers", txCtx).
			Return([]*db.User{
				localUserWithCorrectName, localUserWithCorrectEmail, localUserThatWillBeDeleted, createdLocalUser,
			}, nil).
			Once()

		dbtx.
			On("DeleteUser", txCtx, localUserThatWillBeDeleted.ID).
			Return(nil).
			Once()

		auditLogger.
			On("Logf", ctx, targetIdentifier("user1@example.com"), auditAction(sqlc.AuditActionUsersyncUpdate), "Local user updated: \"user1@example.com\"").
			Return(nil).
			Once()
		auditLogger.
			On("Logf", ctx, targetIdentifier("user2@example.com"), auditAction(sqlc.AuditActionUsersyncCreate), "Local user created: \"user2@example.com\"").
			Return(nil).
			Once()
		auditLogger.
			On("Logf", ctx, targetIdentifier("user3@example.com"), auditAction(sqlc.AuditActionUsersyncUpdate), "Local user updated: \"user3@example.com\"").
			Return(nil).
			Once()
		auditLogger.
			On("Logf", ctx, targetIdentifier("delete-me@example.com"), auditAction(sqlc.AuditActionUsersyncDelete), "Local user deleted: \"delete-me@example.com\"").
			Return(nil).
			Once()

		usersync := usersync.New(database, auditLogger, domain, httpClient)
		err := usersync.Sync(ctx)
		assert.NoError(t, err)
	})
}

func targetIdentifier(identifier string) interface{} {
	return mock.MatchedBy(func(t []auditlogger.Target) bool {
		return t[0].Identifier == identifier
	})
}

func auditAction(action sqlc.AuditAction) interface{} {
	return mock.MatchedBy(func(f auditlogger.Fields) bool {
		return f.Action == action
	})
}

func serialUuid(serial byte) uuid.UUID {
	return uuid.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, serial}
}
