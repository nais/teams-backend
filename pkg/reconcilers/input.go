package reconcilers

import "github.com/nais/console/pkg/dbmodels"

// ReconcilerInput Input for reconcilers
type reconcilerInput struct {
	corr dbmodels.Correlation
	team dbmodels.Team
}

type ReconcilerInput interface {
	GetValues() (dbmodels.Correlation, dbmodels.Team)
}

func NewReconcilerInput(corr dbmodels.Correlation, team dbmodels.Team) ReconcilerInput {
	return reconcilerInput{
		corr: corr,
		team: team,
	}
}

func (in reconcilerInput) GetValues() (dbmodels.Correlation, dbmodels.Team) {
	return in.corr, in.team
}
