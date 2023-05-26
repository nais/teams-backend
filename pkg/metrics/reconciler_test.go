package metrics_test

import (
	"testing"

	"github.com/nais/teams-backend/pkg/metrics"
	"github.com/nais/teams-backend/pkg/sqlc"
)

func Test_MeasureReconcilerDurations(t *testing.T) {
	metrics.MeasureReconcilerDuration(sqlc.ReconcilerNameGithubTeam)
	metrics.MeasureReconcilerDuration("")
}
