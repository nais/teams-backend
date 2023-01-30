package metrics_test

import (
	"testing"

	"github.com/nais/console/pkg/metrics"
	"github.com/nais/console/pkg/sqlc"
)

func Test_MeasureReconcilerDurations(t *testing.T) {
	metrics.MeasureReconcilerDuration(sqlc.ReconcilerNameGithubTeam)
	metrics.MeasureReconcilerDuration("")

	metrics.MeasureReconcileTeamDuration("team-slug")
	metrics.MeasureReconcileTeamDuration("")
}
