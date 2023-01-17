package fixtures

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nais/console/pkg/db"
)

const (
	slug         = "nais-verification"
	purpose      = "A place for NAIS to run verification workloads"
	slackChannel = "#nais-alerts-prod"
)

func CreateNaisVerification(ctx context.Context, database db.Database) error {
	_, err := database.GetTeamBySlug(ctx, slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			_, err = database.CreateTeam(ctx, slug, purpose, slackChannel)
		}
	}

	return err
}
