//go:build run_manually_for_data_migration

package legacy_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/legacy"
	"github.com/nais/console/pkg/roles"
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

			log.Debugf("Fetch team info for %s...", *team.Name)
			members, err := gimp.GroupMembers(yamlteam.AzureID)
			if err != nil {
				return err
			}

			validMembers := make([]*dbmodels.User, 0, len(members))
			for _, member := range members {
				if !strings.HasSuffix(*member.Email, cfg.Google.Domain) {
					log.Warnf("Skip member %s", *member.Email)
					continue
				}
				tx.FirstOrCreate(member, "email = ?", *member.Email)
				log.Debugf("Created user %s", *member.Email)
				validMembers = append(validMembers, member)
			}

			team.Users = validMembers
			tx.Save(team)

			log.Infof("Fetch team administrators for %s...", *team.Name)
			owners, err := gimp.GroupOwners(yamlteam.AzureID)
			for _, owner := range owners {
				if !strings.HasSuffix(*owner.Email, cfg.Google.Domain) {
					log.Warnf("Skip owner %s", *owner.Email)
					continue
				}
				tx.FirstOrCreate(owner, "email = ?", *owner.Email)
				log.Debugf("Created user %s", *owner.Email)
				rb := &dbmodels.RoleBinding{
					RoleID: roles.TeamManagerID,
					TeamID: team.ID,
					UserID: owner.ID,
				}
				tx.Save(rb)
				validMembers = append(validMembers, owner)
			}

			team.Users = validMembers
			tx.Save(team)

			log.Infof("Created %s with %d owners and %d members", *team.Name, len(owners), len(members))

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
