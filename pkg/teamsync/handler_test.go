package teamsync_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nais/console/pkg/slug"

	github_team_reconciler "github.com/nais/console/pkg/reconcilers/github/team"
	nais_deploy_reconciler "github.com/nais/console/pkg/reconcilers/nais/deploy"

	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/teamsync"
)

func TestHandler_ReconcileTeam(t *testing.T) {
	const teamSlug = slug.Slug("my team")

	ctx := context.Background()
	database := db.NewMockDatabase(t)
	cfg := config.Defaults()
	auditLogger := auditlogger.NewMockAuditLogger(t)
	log := logger.NewMockLogger(t)

	t.Run("no reconcilers", func(t *testing.T) {
		log := logger.NewMockLogger(t)
		log.
			On("WithTeamSlug", string(teamSlug)).
			Return(log)

		log.On("Infof", "reconcile team").Once()
		log.On("Debugf", mock.Anything, mock.Anything).Maybe()

		database.
			On("SetLastSuccessfulSyncForTeam", mock.Anything, teamSlug).
			Return(nil).
			Once()

		input := teamsync.Input{
			CorrelationID: uuid.New(),
			TeamSlug:      teamSlug,
		}
		team := &db.Team{
			Team: &sqlc.Team{
				Slug:    teamSlug,
				Purpose: "some purpose",
			},
		}
		database.On("GetTeamBySlug", mock.Anything, teamSlug).Return(team, nil).Once()
		database.On("GetTeamMembers", mock.Anything, teamSlug).Return(nil, nil).Once()

		handler := teamsync.NewHandler(ctx, database, cfg, auditLogger, log)
		handler.Schedule(input)
		handler.Close()
		handler.SyncTeams(ctx)
	})

	t.Run("use reconciler with missing factory", func(t *testing.T) {
		handler := teamsync.NewHandler(ctx, database, cfg, auditLogger, log)
		handler.SetReconcilerFactories(teamsync.ReconcilerFactories{})
		reconciler := db.Reconciler{Reconciler: &sqlc.Reconciler{Name: nais_deploy_reconciler.Name}}
		assert.ErrorContains(t, handler.UseReconciler(reconciler), "missing reconciler factory")
	})

	t.Run("use reconciler with failing factory", func(t *testing.T) {
		err := errors.New("some error")
		handler := teamsync.NewHandler(ctx, database, cfg, auditLogger, log)
		handler.SetReconcilerFactories(teamsync.ReconcilerFactories{
			nais_deploy_reconciler.Name: func(context.Context, db.Database, *config.Config, auditlogger.AuditLogger, logger.Logger) (reconcilers.Reconciler, error) {
				return nil, err
			},
		})
		reconciler := db.Reconciler{Reconciler: &sqlc.Reconciler{Name: nais_deploy_reconciler.Name}}
		assert.ErrorIs(t, handler.UseReconciler(reconciler), err)
	})

	t.Run("multiple reconcilers", func(t *testing.T) {
		log := logger.NewMockLogger(t)
		log.
			On("WithTeamSlug", string(teamSlug)).
			Return(log)
		log.
			On("Infof", "reconcile team").
			Return(nil).
			Once()
		log.
			On("WithSystem", string(azure_group_reconciler.Name)).
			Return(log).
			Once()
		log.
			On("WithSystem", string(github_team_reconciler.Name)).
			Return(log).
			Once()
		log.
			On("WithSystem", string(nais_deploy_reconciler.Name)).
			Return(log).
			Once()
		log.
			On("Debugf", mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "successful reconcile duration")
			}), mock.Anything).
			Return(nil).
			Once()
		log.
			On("Debugf", mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "successful reconcile duration")
			}), mock.Anything).
			Return(nil).
			Once()
		log.
			On("Debugf", mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "successful reconcile duration")
			}), mock.Anything).
			Return(nil).
			Once()
		log.
			On("Debugf", mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "successful reconcile duration")
			}), mock.Anything).
			Return(nil).
			Once()

		database := db.NewMockDatabase(t)
		database.
			On("ClearReconcilerErrorsForTeam", mock.Anything, teamSlug, azure_group_reconciler.Name).
			Return(nil).
			Once()
		database.
			On("ClearReconcilerErrorsForTeam", mock.Anything, teamSlug, github_team_reconciler.Name).
			Return(nil).
			Once()
		database.
			On("ClearReconcilerErrorsForTeam", mock.Anything, teamSlug, nais_deploy_reconciler.Name).
			Return(nil).
			Once()
		database.
			On("SetLastSuccessfulSyncForTeam", mock.Anything, teamSlug).
			Return(nil).
			Once()

		team := &db.Team{
			Team: &sqlc.Team{
				Slug:    teamSlug,
				Purpose: "some purpose",
			},
		}
		input := teamsync.Input{
			CorrelationID: uuid.New(),
			TeamSlug:      teamSlug,
		}
		database.On("GetTeamBySlug", mock.Anything, teamSlug).Return(team, nil).Once()
		database.On("GetTeamMembers", mock.Anything, teamSlug).Return(nil, nil).Once()

		runOrder := 1

		createAzureReconciler := func(context.Context, db.Database, *config.Config, auditlogger.AuditLogger, logger.Logger) (reconcilers.Reconciler, error) {
			reconciler := reconcilers.NewMockReconciler(t)
			reconciler.
				On("Name").
				Return(azure_group_reconciler.Name).
				Once()
			reconciler.
				On("Reconcile", mock.Anything, mock.MatchedBy(func(in reconcilers.Input) bool { return in.Team.Slug == teamSlug })).
				Run(func(args mock.Arguments) {
					assert.Equal(t, 1, runOrder)
					runOrder++
				}).
				Return(nil).
				Once()
			return reconciler, nil
		}
		createGitHubReconciler := func(context.Context, db.Database, *config.Config, auditlogger.AuditLogger, logger.Logger) (reconcilers.Reconciler, error) {
			reconciler := reconcilers.NewMockReconciler(t)
			reconciler.
				On("Name").
				Return(github_team_reconciler.Name).
				Once()
			reconciler.
				On("Reconcile", mock.Anything, mock.MatchedBy(func(in reconcilers.Input) bool { return in.Team.Slug == teamSlug })).
				Run(func(args mock.Arguments) {
					assert.Equal(t, 2, runOrder)
					runOrder++
				}).
				Return(nil).
				Once()
			return reconciler, nil
		}
		createNaisDeployReconciler := func(context.Context, db.Database, *config.Config, auditlogger.AuditLogger, logger.Logger) (reconcilers.Reconciler, error) {
			rec := reconcilers.NewMockReconciler(t)
			rec.
				On("Name").
				Return(nais_deploy_reconciler.Name).
				Once()
			rec.
				On("Reconcile", mock.Anything, mock.MatchedBy(func(in reconcilers.Input) bool { return in.Team.Slug == teamSlug })).
				Run(func(args mock.Arguments) {
					assert.Equal(t, 3, runOrder)
				}).
				Return(nil).
				Once()
			return rec, nil
		}

		handler := teamsync.NewHandler(ctx, database, cfg, auditLogger, log)
		handler.SetReconcilerFactories(teamsync.ReconcilerFactories{
			azure_group_reconciler.Name: createAzureReconciler,
			github_team_reconciler.Name: createGitHubReconciler,
			nais_deploy_reconciler.Name: createNaisDeployReconciler,
		})

		assert.Nil(t, handler.UseReconciler(db.Reconciler{Reconciler: &sqlc.Reconciler{Name: nais_deploy_reconciler.Name, RunOrder: 3}}))
		assert.Nil(t, handler.UseReconciler(db.Reconciler{Reconciler: &sqlc.Reconciler{Name: azure_group_reconciler.Name, RunOrder: 1}}))
		assert.Nil(t, handler.UseReconciler(db.Reconciler{Reconciler: &sqlc.Reconciler{Name: github_team_reconciler.Name, RunOrder: 2}}))
		handler.Schedule(input)
		handler.Close()
		handler.SyncTeams(ctx)
	})

	t.Run("test double schedule ends up with 2 reconciles", func(t *testing.T) {
		log := logger.NewMockLogger(t)
		log.On("WithTeamSlug", mock.Anything).Return(log)
		log.On("Infof", mock.AnythingOfType("string"))
		log.On("Debugf", mock.Anything, mock.Anything)

		input := teamsync.Input{
			CorrelationID: uuid.New(),
			TeamSlug:      teamSlug,
		}
		team := &db.Team{
			Team: &sqlc.Team{
				Slug:    teamSlug,
				Purpose: "some purpose",
			},
		}
		database.
			On("GetTeamBySlug", mock.Anything, teamSlug).
			Return(team, nil).
			Twice()
		database.
			On("GetTeamMembers", mock.Anything, teamSlug).
			Return(nil, nil).
			Twice()
		database.
			On("SetLastSuccessfulSyncForTeam", mock.Anything, teamSlug).
			Return(nil).
			Twice()

		handler := teamsync.NewHandler(ctx, database, cfg, auditLogger, log)
		handler.Schedule(input)
		handler.Schedule(input)
		handler.Close()
		handler.SyncTeams(ctx)
	})
}

func TestHandler_DeleteTeam(t *testing.T) {
	// TODO: Add tests
}
