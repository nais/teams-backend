package graph

import (
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	database       db.Database
	tenantDomain   string
	teamReconciler chan<- reconcilers.Input
	systemName     sqlc.SystemName
	auditLogger    auditlogger.AuditLogger
}

func NewResolver(database db.Database, tenantDomain string, teamReconciler chan<- reconcilers.Input, auditLogger auditlogger.AuditLogger) *Resolver {
	return &Resolver{
		database:       database,
		tenantDomain:   tenantDomain,
		systemName:     sqlc.SystemNameConsole,
		teamReconciler: teamReconciler,
		auditLogger:    auditLogger,
	}
}
