package reconcilers

import "github.com/nais/console/pkg/dbmodels"

// Input Input for reconcilers
type Input struct {
	Corr dbmodels.Correlation
	Team dbmodels.Team
}

// ReconcileTeamInput Input used when triggering the reconciliation of a team
type ReconcileTeamInput struct {
	Team dbmodels.Team
}
