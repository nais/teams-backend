//go:build run_manually

package legacy_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/nais/console/pkg/auditlogger"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/nais/console/pkg/slug"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/legacy"
	log "github.com/sirupsen/logrus"
)

func TestImportTeamsFromLegacyAzure(t *testing.T) {
	const ymlpath = "./teams.yml"
	const jsonpath = "./teams.json"

	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	database, err := setupDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}

	auditLogger := auditlogger.New(database).WithSystemName(sqlc.SystemNameLegacyImporter)

	gimp, err := legacy.NewFromConfig(cfg)
	if err != nil {
		panic(err)
	}

	teams, err := legacy.ReadTeamFiles(ymlpath, jsonpath)
	if err != nil {
		panic(err)
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	users := make(map[string]*db.User)
	err = database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		for _, yamlteam := range teams {
			teamOwners := make(map[string]*db.User, 0)
			teamMembers := make(map[string]*db.User, 0)

			err := slug.Slug(yamlteam.Name).Validate()
			if err != nil {
				log.Warnf("Skip team '%s' as the name is not a valid Console team slug", yamlteam.Name)
				continue
			}

			log.Infof("Fetch team administrators for %s...", yamlteam.Name)
			owners, err := gimp.GroupOwners(yamlteam.AzureID)
			if err != nil {
				return err
			}
			for _, gimpOwner := range owners {
				if !strings.HasSuffix(gimpOwner.Email, cfg.TenantDomain) {
					log.Warnf("Skip owner %s", gimpOwner.Email)
					continue
				}

				if _, exists := users[gimpOwner.Email]; !exists {
					owner, err := dbtx.GetUserByEmail(ctx, gimpOwner.Email)
					if err != nil {
						owner, err = dbtx.AddUser(ctx, gimpOwner.Name, gimpOwner.Email)
						if err != nil {
							return err
						}

						err = auditLogger.Logf(ctx, auditlogger.Fields{
							Action:          sqlc.AuditActionLegacyImporterUserCreate,
							CorrelationID:   correlationID,
							TargetUserEmail: &owner.Email,
						}, "created user")
						if err != nil {
							return err
						}
					}
					users[gimpOwner.Email] = owner
				}
				teamOwners[gimpOwner.Email] = users[gimpOwner.Email]
			}

			log.Infof("Fetch team members for %s...", yamlteam.Name)
			members, err := gimp.GroupMembers(yamlteam.AzureID)
			if err != nil {
				return err
			}
			for _, gimpMember := range members {
				if !strings.HasSuffix(gimpMember.Email, cfg.TenantDomain) {
					log.Warnf("Skip member %s", gimpMember.Email)
					continue
				}

				// check if member is already owner, if so, no need to add it to the member list as well
				if _, isOwner := teamOwners[gimpMember.Email]; isOwner {
					continue
				}

				if _, exists := users[gimpMember.Email]; !exists {
					member, err := dbtx.GetUserByEmail(ctx, gimpMember.Email)
					if err != nil {
						member, err = dbtx.AddUser(ctx, gimpMember.Name, gimpMember.Email)
						if err != nil {
							return err
						}

						err = auditLogger.Logf(ctx, auditlogger.Fields{
							Action:          sqlc.AuditActionLegacyImporterUserCreate,
							CorrelationID:   correlationID,
							TargetUserEmail: &member.Email,
						}, "created user")
						if err != nil {
							return err
						}
					}
					users[gimpMember.Email] = member
				}
				teamMembers[gimpMember.Email] = users[gimpMember.Email]
			}

			if len(teamOwners) == 0 && len(teamMembers) == 0 {
				log.Warnf("The Azure Group '%s' has no members or administrators, skip creation of the team in Console", yamlteam.Name)
				continue
			}

			if len(teamOwners) == 0 {
				_, user := first(teamMembers)
				log.Infof("The Azure Group '%s' has no administrators, setting the first member as owner for the Console team: '%s'", yamlteam.Name, user.Email)
				teamOwners[user.Email] = user
				delete(teamMembers, user.Email)
			}

			_, owner := first(teamOwners)
			convertedTeam := yamlteam.Convert()

			team, err := dbtx.GetTeamBySlug(ctx, convertedTeam.Slug)
			if err != nil {
				team, err = dbtx.AddTeam(ctx, convertedTeam.Name, convertedTeam.Slug, &convertedTeam.Purpose.String, owner.ID)
				if err != nil {
					return err
				}

				// TODO: Store metadata

				err = auditLogger.Logf(ctx, auditlogger.Fields{
					Action:         sqlc.AuditActionLegacyImporterTeamCreate,
					CorrelationID:  correlationID,
					TargetTeamSlug: &team.Slug,
				}, "add team")
				if err != nil {
					return err
				}
			}

			for _, user := range teamOwners {
				err := dbtx.SetTeamMemberRole(ctx, user.ID, team.ID, sqlc.RoleNameTeamowner)
				if err != nil {
					return err
				}

				err = auditLogger.Logf(ctx, auditlogger.Fields{
					Action:          sqlc.AuditActionLegacyImporterTeamAddOwner,
					CorrelationID:   correlationID,
					TargetTeamSlug:  &team.Slug,
					TargetUserEmail: &user.Email,
				}, "add team owner")
				if err != nil {
					return err
				}
			}

			for _, user := range teamMembers {
				err := dbtx.SetTeamMemberRole(ctx, user.ID, team.ID, sqlc.RoleNameTeammember)
				if err != nil {
					return err
				}

				err = auditLogger.Logf(ctx, auditlogger.Fields{
					Action:          sqlc.AuditActionLegacyImporterTeamAddMember,
					CorrelationID:   correlationID,
					TargetTeamSlug:  &team.Slug,
					TargetUserEmail: &user.Email,
				}, "add team member")
				if err != nil {
					return err
				}
			}

			log.Infof("Created team '%s' with %d owners and %d members", team.Name, len(teamOwners), len(teamMembers))
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	log.Infof("Done")
}

func setupDatabase(ctx context.Context, dbUrl string) (db.Database, error) {
	dbc, err := pgxpool.Connect(ctx, dbUrl)
	if err != nil {
		return nil, err
	}

	err = db.Migrate(dbc.Config().ConnString())
	if err != nil {
		return nil, err
	}

	queries := db.Wrap(sqlc.New(dbc), dbc)
	return db.NewDatabase(queries), nil
}

func first(users map[string]*db.User) (string, *db.User) {
	var k string
	var u *db.User

	for k, u = range users {
		break
	}

	return k, u
}
