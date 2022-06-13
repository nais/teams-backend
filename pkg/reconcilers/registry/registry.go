package registry

import (
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	"gorm.io/gorm"
)

type ReconcilerInitializer func(*gorm.DB, *config.Config, dbmodels.System, auditlogger.AuditLogger) (reconcilers.Reconciler, error)

var recs = make(map[string]ReconcilerInitializer)

func Register(name string, init ReconcilerInitializer) {
	recs[name] = init
}

func Reconcilers() map[string]ReconcilerInitializer {
	return recs
}
