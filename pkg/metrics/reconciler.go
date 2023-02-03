package metrics

import (
	"net/http"
	"strconv"

	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "nais"
	subsystem = "console"
)

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
		Help:      "Number of reconcile runs, labeled with their ID and run state",
	}, []string{labelReconciler, labelState})

	externalCalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "external_api_requests",
		Help:      "Number of API requests done to external systems, labeled with status code and system name",
	}, []string{labelSystem, labelStatusCode})

	pendingTeams = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pending_teams",
		Help:      "How many teams currently pending reconciliation with external systems",
	})

	reconcilerMaxAttemptsExhaustion = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "reconcile_max_attempts_exhaustion",
		Help:      "Number of times a team has exhausted all its sync attempts",
	})

	reconcilerDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "reconciler_duration",
		Help:      "Duration of a specific reconciler, regardless of team.",
		Buckets:   prometheus.LinearBuckets(0, .5, 20),
	}, []string{"reconciler"})

	reconcileTeamDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "reconcile_team_duration",
		Help:      "Reconcile duration of a specific team.",
		Buckets:   prometheus.LinearBuckets(0, 2, 20),
	}, []string{"team"})
)

func IncReconcilerCounter(name sqlc.ReconcilerName, state ReconcilerState) {
	labels := prometheus.Labels{
		labelReconciler: string(name),
		labelState:      string(state),
	}
	reconcilerCounter.With(labels).Inc()
}

func IncReconcilerMaxAttemptsExhaustion() {
	reconcilerMaxAttemptsExhaustion.Inc()
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

func SetPendingTeamCount(numTeams int) {
	pendingTeams.Set(float64(numTeams))
}

func MeasureReconcilerDuration(reconciler sqlc.ReconcilerName) *prometheus.Timer {
	return prometheus.NewTimer(reconcilerDuration.WithLabelValues(string(reconciler)))
}

func MeasureReconcileTeamDuration(teamSlug slug.Slug) *prometheus.Timer {
	return prometheus.NewTimer(reconcileTeamDuration.WithLabelValues(string(teamSlug)))
}
