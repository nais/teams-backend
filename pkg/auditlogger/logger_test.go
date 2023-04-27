package auditlogger_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Logf(t *testing.T) {
	ctx := context.Background()
	database := db.NewMockDatabase(t)
	msg := "some message"
	system := sqlc.SystemNameConsole

	t.Run("missing system name", func(t *testing.T) {
		log := logger.NewMockLogger(t)
		log.
			On("Errorf", mock.MatchedBy(func(msg string) bool {
				return strings.Contains(msg, "missing or invalid system name")
			})).
			Once()
		auditLogger := auditlogger.New(log)
		auditLogger.Logf(ctx, database, []auditlogger.Target{}, auditlogger.Fields{}, msg)
	})

	t.Run("missing audit action", func(t *testing.T) {
		log := logger.NewMockLogger(t)
		log.
			On("WithSystem", string(system)).
			Return(log).
			Once()
		log.
			On("Errorf", mock.MatchedBy(func(msg string) bool {
				return strings.Contains(msg, "missing or invalid audit action")
			})).
			Once()

		auditLogger := auditlogger.New(log)
		auditLogger.
			WithSystemName(system).
			Logf(ctx, database, []auditlogger.Target{}, auditlogger.Fields{}, msg)
	})

	t.Run("does not do anything without targets", func(t *testing.T) {
		log := logger.NewMockLogger(t)
		log.
			On("WithSystem", string(system)).
			Return(log).
			Once()
		auditLogger := auditlogger.New(log)
		fields := auditlogger.Fields{
			Action: sqlc.AuditActionAzureGroupAddMember,
		}
		auditLogger.
			WithSystemName(system).
			Logf(ctx, database, []auditlogger.Target{}, fields, msg)
	})

	t.Run("log with target and all fields", func(t *testing.T) {
		testLogger, hook := test.NewNullLogger()
		_ = hook

		userEmail := "mail@example.com"
		teamSlug := slug.Slug("team-slug")
		reconcilerName := sqlc.ReconcilerNameGithubTeam
		systemName := sqlc.SystemNameGithubTeam
		actorIdentity := "actor"
		action := sqlc.AuditActionAzureGroupAddMember

		correlationID := uuid.New()
		targets := []auditlogger.Target{
			auditlogger.UserTarget(userEmail),
			auditlogger.TeamTarget(teamSlug),
			auditlogger.ReconcilerTarget(reconcilerName),
			auditlogger.SystemTarget(systemName),
		}

		log := logger.NewMockLogger(t)
		log.On("WithSystem", string(system)).Return(log).Once()
		log.On("WithActor", actorIdentity).Return(log).Times(len(targets))
		log.On("WithUser", userEmail).Return(log).Once()
		log.On("WithTeamSlug", string(teamSlug)).Return(log).Once()
		log.On("WithReconciler", string(reconcilerName)).Return(log).Once()
		log.
			On("WithFields", mock.MatchedBy(func(f logrus.Fields) bool {
				return f["action"] == action &&
					f["actor"] == actorIdentity &&
					f["correlation_id"] == correlationID &&
					f["target_type"] == sqlc.AuditLogsTargetTypeUser
			})).
			Return(&logrus.Entry{Logger: testLogger}).
			Once()
		log.
			On("WithFields", mock.MatchedBy(func(f logrus.Fields) bool {
				return f["action"] == action &&
					f["actor"] == actorIdentity &&
					f["correlation_id"] == correlationID &&
					f["target_type"] == sqlc.AuditLogsTargetTypeTeam
			})).
			Return(&logrus.Entry{Logger: testLogger}).
			Once()
		log.
			On("WithFields", mock.MatchedBy(func(f logrus.Fields) bool {
				return f["action"] == action &&
					f["actor"] == actorIdentity &&
					f["correlation_id"] == correlationID &&
					f["target_type"] == sqlc.AuditLogsTargetTypeReconciler
			})).
			Return(&logrus.Entry{Logger: testLogger}).
			Once()
		log.
			On("WithFields", mock.MatchedBy(func(f logrus.Fields) bool {
				return f["action"] == action &&
					f["actor"] == actorIdentity &&
					f["correlation_id"] == correlationID &&
					f["target_type"] == sqlc.AuditLogsTargetTypeSystem
			})).
			Return(&logrus.Entry{Logger: testLogger}).
			Once()

		auditLogger := auditlogger.New(log)

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
			On("CreateAuditLogEntry", ctx, correlationID, system, &actorIdentity, sqlc.AuditLogsTargetTypeUser, userEmail, action, msg).
			Return(nil).
			Once()
		database.
			On("CreateAuditLogEntry", ctx, correlationID, system, &actorIdentity, sqlc.AuditLogsTargetTypeTeam, string(teamSlug), action, msg).
			Return(nil).
			Once()
		database.
			On("CreateAuditLogEntry", ctx, correlationID, system, &actorIdentity, sqlc.AuditLogsTargetTypeReconciler, string(reconcilerName), action, msg).
			Return(nil).
			Once()
		database.
			On("CreateAuditLogEntry", ctx, correlationID, system, &actorIdentity, sqlc.AuditLogsTargetTypeSystem, string(systemName), action, msg).
			Return(nil).
			Once()
		auditLogger.
			WithSystemName(system).
			Logf(ctx, database, targets, fields, msg)

		assert.Len(t, hook.Entries, len(targets))
		assert.Equal(t, msg, hook.Entries[0].Message)
		assert.Equal(t, msg, hook.Entries[1].Message)
		assert.Equal(t, msg, hook.Entries[2].Message)
		assert.Equal(t, msg, hook.Entries[3].Message)
	})
}
