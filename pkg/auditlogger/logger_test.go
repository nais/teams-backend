package auditlogger_test

import (
	"context"
	"strings"
	"testing"

	"github.com/nais/teams-backend/pkg/types"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/authz"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Logf(t *testing.T) {
	ctx := context.Background()
	database := db.NewMockDatabase(t)
	msg := "some message"
	componentName := types.ComponentNameConsole

	t.Run("missing audit action", func(t *testing.T) {
		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", componentName).
			Return(log).
			Once()
		log.
			On("Errorf", mock.MatchedBy(func(msg string) bool {
				return strings.Contains(msg, "missing or invalid audit action")
			})).
			Once()

		auditlogger.
			New(database, componentName, log).
			Logf(ctx, []auditlogger.Target{}, auditlogger.Fields{}, msg)
	})

	t.Run("does not do anything without targets", func(t *testing.T) {
		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", componentName).
			Return(log).
			Once()

		fields := auditlogger.Fields{
			Action: types.AuditActionAzureGroupAddMember,
		}
		auditlogger.
			New(database, componentName, log).
			Logf(ctx, []auditlogger.Target{}, fields, msg)
	})

	t.Run("log with target and all fields", func(t *testing.T) {
		testLogger, hook := test.NewNullLogger()
		_ = hook

		userEmail := "mail@example.com"
		teamSlug := slug.Slug("team-slug")
		reconcilerName := sqlc.ReconcilerNameGithubTeam
		componentName := types.ComponentNameGithubTeam
		actorIdentity := "actor"
		action := types.AuditActionAzureGroupAddMember

		correlationID := uuid.New()
		targets := []auditlogger.Target{
			auditlogger.UserTarget(userEmail),
			auditlogger.TeamTarget(teamSlug),
			auditlogger.ReconcilerTarget(reconcilerName),
			auditlogger.ComponentTarget(componentName),
		}

		log := logger.NewMockLogger(t)
		log.On("WithComponent", componentName).Return(log).Once()
		log.On("WithActor", actorIdentity).Return(log).Times(len(targets))
		log.On("WithUser", userEmail).Return(log).Once()
		log.On("WithTeamSlug", string(teamSlug)).Return(log).Once()
		log.On("WithReconciler", string(reconcilerName)).Return(log).Once()
		log.
			On("WithFields", mock.MatchedBy(func(f logrus.Fields) bool {
				return f["action"] == action &&
					f["actor"] == actorIdentity &&
					f["correlation_id"] == correlationID &&
					f["target_type"] == types.AuditLogsTargetTypeUser
			})).
			Return(&logrus.Entry{Logger: testLogger}).
			Once()
		log.
			On("WithFields", mock.MatchedBy(func(f logrus.Fields) bool {
				return f["action"] == action &&
					f["actor"] == actorIdentity &&
					f["correlation_id"] == correlationID &&
					f["target_type"] == types.AuditLogsTargetTypeTeam
			})).
			Return(&logrus.Entry{Logger: testLogger}).
			Once()
		log.
			On("WithFields", mock.MatchedBy(func(f logrus.Fields) bool {
				return f["action"] == action &&
					f["actor"] == actorIdentity &&
					f["correlation_id"] == correlationID &&
					f["target_type"] == types.AuditLogsTargetTypeReconciler
			})).
			Return(&logrus.Entry{Logger: testLogger}).
			Once()
		log.
			On("WithFields", mock.MatchedBy(func(f logrus.Fields) bool {
				return f["action"] == action &&
					f["actor"] == actorIdentity &&
					f["correlation_id"] == correlationID &&
					f["target_type"] == types.AuditLogsTargetTypeSystem
			})).
			Return(&logrus.Entry{Logger: testLogger}).
			Once()

		authenticatedUser := db.NewMockAuthenticatedUser(t)
		authenticatedUser.On("Identity").Return(actorIdentity).Once()

		fields := auditlogger.Fields{
			Action: action,
			Actor: &authz.Actor{
				User: authenticatedUser,
			},
			CorrelationID: correlationID,
		}

		database := db.NewMockDatabase(t)
		database.
			On("CreateAuditLogEntry", ctx, correlationID, componentName, &actorIdentity, types.AuditLogsTargetTypeUser, userEmail, action, msg).
			Return(nil).
			Once()
		database.
			On("CreateAuditLogEntry", ctx, correlationID, componentName, &actorIdentity, types.AuditLogsTargetTypeTeam, string(teamSlug), action, msg).
			Return(nil).
			Once()
		database.
			On("CreateAuditLogEntry", ctx, correlationID, componentName, &actorIdentity, types.AuditLogsTargetTypeReconciler, string(reconcilerName), action, msg).
			Return(nil).
			Once()
		database.
			On("CreateAuditLogEntry", ctx, correlationID, componentName, &actorIdentity, types.AuditLogsTargetTypeSystem, string(componentName), action, msg).
			Return(nil).
			Once()

		auditlogger.
			New(database, componentName, log).
			Logf(ctx, targets, fields, msg)

		assert.Len(t, hook.Entries, len(targets))
		assert.Equal(t, msg, hook.Entries[0].Message)
		assert.Equal(t, msg, hook.Entries[1].Message)
		assert.Equal(t, msg, hook.Entries[2].Message)
		assert.Equal(t, msg, hook.Entries[3].Message)
	})
}
