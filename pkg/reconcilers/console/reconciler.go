package console_reconciler

import (
	"context"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
)

type consoleReconciler struct{}

const (
	Name                   = sqlc.SystemNameConsole
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

func New() *consoleReconciler {
	return &consoleReconciler{}
}

func NewFromConfig(_ db.Database, _ *config.Config, _ auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	return New(), nil
}

func (r *consoleReconciler) Reconcile(_ context.Context, _ reconcilers.Input) error {
	return nil
}

func (r *consoleReconciler) Name() sqlc.SystemName {
	return Name
}
