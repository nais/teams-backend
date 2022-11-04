package fixtures

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nais/console/pkg/db"
)

const (
	Slug    = "nais-verification"
	Purpose = "A place for NAIS to run verification workloads"
)

func CreateNaisVerification(ctx context.Context, database db.Database) error {
	_, err := database.GetTeamBySlug(ctx, Slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			_, err = database.CreateTeam(ctx, Slug, Purpose)
		}
	}

	return err
}
