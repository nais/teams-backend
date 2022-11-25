package metrics

import (
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/sqlc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

func IncReconcilerCounter(name sqlc.ReconcilerName, state ReconcilerState, log logger.Logger) {
	labels := prometheus.Labels{
		"reconciler": string(name),
		"state":      string(state),
	}
	m, err := reconcilerCounterMetric.GetMetricWith(labels)
	if err != nil {
		// TODO wut system
		log.WithSystem("metrics").WithError(err).Error("get metric with labels: %+v", labels)
		log.Errorf("failed getting metric: %v", err)
	}
	m.Inc()
}
