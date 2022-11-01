package fixtures

import (
	"context"
	"github.com/nais/console/pkg/db"
)

const (
	Slug    = "nais"
	Purpose = "A place for NAIS to run non-critical workloads"
)

func CreateNaisTeam(ctx context.Context, database db.Database) error {
	_, err := database.CreateTeam(ctx, Slug, Purpose)
	if err != nil {
		return err
	}

	return nil
}
