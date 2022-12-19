package metrics

import (
	"net/http"
	"strconv"

	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/sqlc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "nais"
const subsystem = "console"

type ReconcilerState string

const (
	ReconcilerStateStarted    ReconcilerState = "started"
	ReconcilerStateFailed     ReconcilerState = "failed"
	ReconcilerStateSuccessful ReconcilerState = "successful"
)

const (
	labelReconciler = "reconciler"
	labelState      = "state"
	labelSystem     = "system"
	labelStatusCode = "status_code"
)

var (
	reconcilerCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "reconciles",
		Help:      "write a proper description",
	}, []string{labelReconciler, labelState})

	externalCalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "external_api_requests",
		Help:      "Number of API requests done to external systems, labeled with status code and system name",
	}, []string{labelSystem, labelStatusCode})
)

func IncReconcilerCounter(name sqlc.ReconcilerName, state ReconcilerState, log logger.Logger) {
	labels := prometheus.Labels{
		labelReconciler: string(name),
		labelState:      string(state),
	}
	reconcilerCounter.With(labels).Inc()
}

func IncExternalHTTPCalls(systemName string, resp *http.Response, err error) {
	var statusCode int
	if resp != nil {
		statusCode = resp.StatusCode
	} else if err != nil {
		statusCode = 1
	}
	labels := prometheus.Labels{
		labelSystem:     systemName,
		labelStatusCode: strconv.Itoa(statusCode),
	}
	externalCalls.With(labels).Inc()
}

func IncExternalCalls(systemName string, statusCode int) {
	labels := prometheus.Labels{
		labelSystem:     systemName,
		labelStatusCode: strconv.Itoa(statusCode),
	}
	externalCalls.With(labels).Inc()
}

func IncExternalCallsByError(systemName string, err error) {
	var statusCode int
	if err != nil {
		statusCode = 1
	}
	labels := prometheus.Labels{
		labelSystem:     systemName,
		labelStatusCode: strconv.Itoa(statusCode),
	}
	externalCalls.With(labels).Inc()
}
