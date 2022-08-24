//go:build run_manually_for_data_migration

package legacy_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/google/uuid"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/legacy"
	log "github.com/sirupsen/logrus"
)

func TestImportTeamsFromLegacyAzure(t *testing.T) {
	const ymlpath = "/Users/kimt/src/navikt/teams/teams.yml"
	const jsonpath = "/Users/kimt/src/navikt/teams/teams.json"

	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	database, err := setupDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}

	gimp, err := legacy.NewFromConfig(cfg)
	if err != nil {
		panic(err)
	}

	teams, err := legacy.ReadTeamFiles(ymlpath, jsonpath)
	if err != nil {
		panic(err)
	}

	dbteams := make([]db.Team, 0, len(teams))

	err = database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		for _, yamlteam := range teams {
			team := yamlteam.Convert()

			log.Debugf("Fetch team info for %s...", team.Name)
			members, err := gimp.GroupMembers(yamlteam.AzureID)
			if err != nil {
				return err
			}

			for _, gimpMember := range members {
				if !strings.HasSuffix(gimpMember.Email, cfg.TenantDomain) {
					log.Warnf("Skip member %s", gimpMember.Email)
					continue
				}

				member, err := dbtx.GetUserByEmail(ctx, gimpMember.Email)
				if err != nil {
					member, err = dbtx.AddUser(ctx, gimpMember.Name, gimpMember.Email)
					if err != nil {
						return err
					}
				}
				log.Debugf("Created user %s", member.Email)

				err = dbtx.AddUsersToTeam(ctx, []uuid.UUID{member.ID}, team.ID)
				if err != nil {
					return err
				}
			}

			log.Infof("Fetch team administrators for %s...", team.Name)
			owners, err := gimp.GroupOwners(yamlteam.AzureID)
			for _, gimpOwner := range owners {
				if !strings.HasSuffix(gimpOwner.Email, cfg.TenantDomain) {
					log.Warnf("Skip owner %s", gimpOwner.Email)
					continue
				}

				owner, err := dbtx.GetUserByEmail(ctx, gimpOwner.Email)
				if err != nil {
					owner, err = dbtx.AddUser(ctx, gimpOwner.Name, gimpOwner.Email)
					if err != nil {
						return err
					}
				}
				log.Debugf("Created user %s", owner.Email)

				err = dbtx.AddUsersToTeam(ctx, []uuid.UUID{owner.ID}, team.ID)
				if err != nil {
					return err
				}

				err = dbtx.SetTeamMembersRole(ctx, []uuid.UUID{owner.ID}, team.ID, sqlc.RoleNameTeamowner)
				if err != nil {
					return err
				}

			}

			log.Infof("Created %s with %d owners and %d members", team.Name, len(owners), len(members))

			dbteams = append(dbteams, *team)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(dbteams)
}

func setupDatabase(ctx context.Context, dbUrl string) (db.Database, error) {
	dbc, err := pgxpool.Connect(ctx, dbUrl)
	if err != nil {
		return nil, err
	}

	queries := db.Wrap(sqlc.New(dbc), dbc)
	return db.NewDatabase(queries, dbc), nil
}
