package registry

import (
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
)

type ReconcilerInitializer func(*config.Config, auditlogger.Logger) (reconcilers.Reconciler, error)

var recs = make(map[string]ReconcilerInitializer)

func Register(name string, init ReconcilerInitializer) {
	recs[name] = init
}

func Reconcilers() map[string]ReconcilerInitializer {
	return recs
}
