//go:build run_manually_for_data_migration

package legacy_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/legacy"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestImportTeamsFromLegacyAzure(t *testing.T) {
	const ymlpath = "/Users/kimt/src/navikt/teams/teams.yml"
	const jsonpath = "/Users/kimt/src/navikt/teams/teams.json"

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	db, err := setupDatabase(cfg)
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

	dbteams := make([]*dbmodels.Team, 0, len(teams))

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, yamlteam := range teams {
			if tx.Error != nil {
				return tx.Error
			}
			team := yamlteam.Convert()
			log.Infof("Fetch team info for %s...", *team.Name)
			members, err := gimp.GroupMembers(yamlteam.AzureID)
			if err != nil {
				return err
			}
			for _, member := range members {
				tx.FirstOrCreate(member, "email = ?", member.Email)
			}
			log.Infof("Created %s with %d members", *team.Name, len(members))
			team.Users = members
			tx.Save(team)
			dbteams = append(dbteams, team)
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

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	log.Infof("Connecting to database...")
	db, err := gorm.Open(
		postgres.New(
			postgres.Config{
				DSN:                  cfg.DatabaseURL,
				PreferSimpleProtocol: true, // disables implicit prepared statement usage
			},
		),
		&gorm.Config{},
	)
	if err != nil {
		return nil, err
	}
	log.Infof("Successfully connected to database.")

	// uuid-ossp is needed for PostgreSQL to generate UUIDs as primary keys
	tx := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if tx.Error != nil {
		return nil, fmt.Errorf("install postgres uuid extension: %w", tx.Error)
	}

	log.Infof("Migrating database schema...")
	err = dbmodels.Migrate(db)
	if err != nil {
		return nil, err
	}

	log.Infof("Successfully migrated database schema.")
	return db, nil
}
