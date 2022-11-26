package logger_test

import (
	"io"
	"testing"

	"github.com/nais/console/pkg/sqlc"

	"github.com/sirupsen/logrus/hooks/test"

	"github.com/stretchr/testify/assert"

	"github.com/nais/console/pkg/logger"
)

const (
	logFormat = "text"
	logLevel  = "DEBUG"

	actorKey      = "actor"
	reconcilerKey = "reconciler"
	systemKey     = "system"
	teamKey       = "team"
	userKey       = "user"

	actor      = "actor@example.com"
	reconciler = "nais:namespace"
	baseSystem = sqlc.SystemNameConsole
	system     = "usersync"
	teamSlug   = "team-slug"
	user       = "user@example.com"
)

func Test_logger_GetLogger(t *testing.T) {
	t.Run("invalid format", func(t *testing.T) {
		_, err := logger.GetLogger("format", "DEBUG")
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid log format: format")
	})

	t.Run("invalid level", func(t *testing.T) {
		_, err := logger.GetLogger("json", "foobar")
		assert.Error(t, err)
		assert.EqualError(t, err, `not a valid logrus Level: "foobar"`)
	})
}

func Test_logger_WithFields(t *testing.T) {
	base, _ := logger.GetLogger(logFormat, logLevel)
	internalLogger := base.GetInternalLogger()
	internalLogger.Out = io.Discard // don't need to see the actual logs
	logHook := test.NewLocal(internalLogger)

	t.Run("base logger", func(t *testing.T) {
		base.Info("some info")
		fields := logHook.LastEntry().Data
		assert.Contains(t, fields, systemKey)
		assert.Equal(t, baseSystem, fields[systemKey])
	})

	t.Run("actor logger", func(t *testing.T) {
		base.WithActor(actor).Warn("some warning")
		fields := logHook.LastEntry().Data
		assert.Contains(t, fields, actorKey)
		assert.Equal(t, actor, fields[actorKey])
	})

	t.Run("reconciler logger", func(t *testing.T) {
		base.WithReconciler(reconciler).Error("some error")
		fields := logHook.LastEntry().Data
		assert.Contains(t, fields, reconcilerKey)
		assert.Equal(t, reconciler, fields[reconcilerKey])
	})

	t.Run("system logger", func(t *testing.T) {
		base.WithSystem(system).Error("some error")
		fields := logHook.LastEntry().Data
		assert.Contains(t, fields, systemKey)
		assert.Equal(t, system, fields[systemKey])
	})

	t.Run("team logger", func(t *testing.T) {
		base.WithTeamSlug(teamSlug).Info("some info")
		fields := logHook.LastEntry().Data
		assert.Contains(t, fields, teamKey)
		assert.Equal(t, teamSlug, fields[teamKey])
	})

	t.Run("user logger", func(t *testing.T) {
		base.WithUser(user).Debug("some debug")
		fields := logHook.LastEntry().Data
		assert.Contains(t, fields, userKey)
		assert.Equal(t, user, fields[userKey])
	})

	t.Run("multiple loggers", func(t *testing.T) {
		actorLogger := base.WithActor(actor)
		reconcilerLogger := actorLogger.WithReconciler(reconciler)
		systemLogger := reconcilerLogger.WithSystem(system)
		teamLogger := systemLogger.WithTeamSlug(teamSlug)
		userLogger := teamLogger.WithUser(user)

		actorLogger.Info("actor info")
		actorEntry := logHook.LastEntry()
		reconcilerLogger.Info("reconciler info")
		reconcilerEntry := logHook.LastEntry()
		systemLogger.Info("system info")
		systemEntry := logHook.LastEntry()
		teamLogger.Info("team info")
		teamEntry := logHook.LastEntry()
		userLogger.Info("user info")
		userEntry := logHook.LastEntry()

		assert.NotContains(t, actorEntry.Data, reconcilerKey)
		assert.Equal(t, reconcilerEntry.Data[systemKey], baseSystem)
		assert.NotContains(t, systemEntry.Data, teamKey)
		assert.NotContains(t, teamEntry.Data, userKey)
		assert.Contains(t, userEntry.Data, actorKey)
		assert.Contains(t, userEntry.Data, reconcilerKey)
		assert.Contains(t, userEntry.Data, systemKey)
		assert.Contains(t, userEntry.Data, teamKey)
		assert.Contains(t, userEntry.Data, userKey)
	})
}
