package registry

import (
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReconcilerFactory func(*gorm.DB, *config.Config, sqlc.System, auditlogger.AuditLogger) (reconcilers.Reconciler, error)

type ReconcilerInitializer struct {
	Name    string
	Factory ReconcilerFactory
}

var recs = make([]ReconcilerInitializer, 0)
var recNames = make(map[string]bool)

func Register(name string, init ReconcilerFactory) {
	if _, exists := recNames[name]; exists {
		log.Warnf("reconciler '%s' has already been registered", name)
		return
	}

	recs = append(recs, ReconcilerInitializer{
		Name:    name,
		Factory: init,
	})
	recNames[name] = true
}

func Reconcilers() []ReconcilerInitializer {
	return recs
}
