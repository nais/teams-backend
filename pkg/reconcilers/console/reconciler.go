package console_reconciler

import (
	"context"
	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"gorm.io/gorm"
)

type consoleReconciler struct {
	system sqlc.System
}

const (
	Name                   = "console"
	OpAddTeamMember        = "console:team:add-member"
	OpAddTeamOwner         = "console:team:add-owner"
	OpCreateApiKey         = "console:api-key:create"
	OpCreateServiceAccount = "console:service-account:create"
	OpCreateTeam           = "console:team:create"
	OpDeleteApiKey         = "console:api-key:delete"
	OpDeleteServiceAccount = "console:service-account:delete"
	OpRemoveTeamMember     = "console:team:add-member"
	OpSetTeamMemberRole    = "console:team:set-member-role"
	OpSyncTeam             = "console:team:sync"
	OpUpdateServiceAccount = "console:service-account:update"
	OpUpdateTeam           = "console:team:update"
)

func New(system sqlc.System) *consoleReconciler {
	return &consoleReconciler{
		system: system,
	}
}

func NewFromConfig(_ *gorm.DB, _ *config.Config, system sqlc.System, _ auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	return New(system), nil
}

func (r *consoleReconciler) Reconcile(_ context.Context, _ reconcilers.Input) error {
	return nil
}

func (r *consoleReconciler) System() sqlc.System {
	return r.system
}
