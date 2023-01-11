package fixtures

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"
	"github.com/sirupsen/logrus"
)

type EnableableReconciler sqlc.ReconcilerName

var enableableReconcilers = []sqlc.ReconcilerName{
	sqlc.ReconcilerNameGoogleGcpProject,
	sqlc.ReconcilerNameGoogleWorkspaceAdmin,
	sqlc.ReconcilerNameNaisDeploy,
	sqlc.ReconcilerNameNaisNamespace,
}

func (e *EnableableReconciler) Decode(s string) error {
	for _, valid := range enableableReconcilers {
		if sqlc.ReconcilerName(s) == valid {
			*e = EnableableReconciler(s)
			return nil
		}
	}

	return fmt.Errorf("reconciler %q cannot be enabled on first run", s)
}

func SetupDefaultReconcilers(ctx context.Context, log *logrus.Entry, reconcilers []EnableableReconciler, database db.Database) error {
	if len(reconcilers) == 0 {
		log.Infof("CONSOLE_FIRST_RUN_ENABLE_RECONCILERS not set or empty - not enabling any reconcilers")
		return nil
	}

	log.Infof("enablling reconcilers: %v", reconcilers)
	for _, reconciler := range reconcilers {
		_, err := database.EnableReconciler(ctx, sqlc.ReconcilerName(reconciler))
		if err != nil {
			return err
		}
	}

	return nil
}