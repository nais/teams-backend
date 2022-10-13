package metrics

import (
	"github.com/nais/console/pkg/sqlc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

const subsystem = "reconciler"

type ReconcilerState string

const (
	ReconcilerStateStarted    ReconcilerState = "started"
	ReconcilerStateFailed     ReconcilerState = "failed"
	ReconcilerStateSuccessful ReconcilerState = "successful"
)

var reconcilerCounterMetric = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "reconciles",
}, []string{"reconciler", "state"})

func IncReconcilerCounter(name sqlc.ReconcilerName, state ReconcilerState) {
	m, err := reconcilerCounterMetric.GetMetricWithLabelValues(string(name), string(state))
	if err != nil {
		log.Errorf("failed getting metric: %v", err)
	}
	m.Inc()
}
