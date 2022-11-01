package fixtures

import (
	"context"
	"github.com/nais/console/pkg/db"
)

const (
	Slug    = "nais-verification"
	Purpose = "A place for NAIS to run verification workloads"
)

func CreateNaisVerification(ctx context.Context, database db.Database) error {
	_, err := database.CreateTeam(ctx, Slug, Purpose)
	if err != nil {
		return err
	}

	return nil
}
