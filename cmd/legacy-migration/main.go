package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/legacy"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := run()
	if err != nil {
		log.Errorf("fatal: %s", err)
		os.Exit(1)
	}
}

func run() error {
	const ymlpath = "./local/teams.yml"
	const jsonpath = "./local/teams.json"
	const gcpJsonCacheTemplate = "./local/gcp-cache/%s-output.json"

	ctx := context.Background()

	cfg, err := config.NewImporterConfig()
	if err != nil {
		panic(err)
	}

	clusters := map[string]any{"dev": nil, "prod": nil}

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

	gcpCacheDev, err := legacy.ReadGcpTeamCacheFile(fmt.Sprintf(gcpJsonCacheTemplate, "dev"))
	if err != nil {
		panic(err)
	}

	gcpCacheProd, err := legacy.ReadGcpTeamCacheFile(fmt.Sprintf(gcpJsonCacheTemplate, "prod"))
	if err != nil {
		panic(err)
	}

	gcpCache := legacy.MergeGcpTeamCacheFiles(gcpCacheDev, gcpCacheProd)

	correlationID := uuid.New()

	users := make(map[string]*db.User)
	err = database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		auditLogger := auditlogger.New(dbtx).WithSystemName(sqlc.SystemNameLegacyImporter)

		for _, yamlteam := range teams {
			teamOwners := make(map[string]*db.User, 0)
			teamMembers := make(map[string]*db.User, 0)

			log.Infof("Fetch team administrators for %q...", yamlteam.Name)
			owners, err := gimp.GroupOwners(yamlteam.AzureGroupID.String())
			if err != nil {
				log.WithError(err).Errorf("Unable to get team owners for team: %q", yamlteam.Name)
				continue
			}
			for _, gimpOwner := range owners {
				if !strings.HasSuffix(gimpOwner.Email, cfg.TenantDomain) {
					log.Warnf("Skip owner %q for team %q", gimpOwner.Email, yamlteam.Name)
					continue
				}

				if _, exists := users[gimpOwner.Email]; !exists {
					owner, err := dbtx.GetUserByEmail(ctx, gimpOwner.Email)
					if err != nil {
						owner, err = dbtx.CreateUser(ctx, gimpOwner.Name, gimpOwner.Email, gimpOwner.ExternalID)
						if err != nil {
							return err
						}

						targets := []auditlogger.Target{
							auditlogger.UserTarget(owner.Email),
						}
						fields := auditlogger.Fields{
							Action:        sqlc.AuditActionLegacyImporterUserCreate,
							CorrelationID: correlationID,
						}
						err = auditLogger.Logf(ctx, targets, fields, "Imported user from Azure AD")
						if err != nil {
							return err
						}
					}
					users[gimpOwner.Email] = owner
				}
				teamOwners[gimpOwner.Email] = users[gimpOwner.Email]
			}

			log.Infof("Fetch team members for %s...", yamlteam.Name)
			members, err := gimp.GroupMembers(yamlteam.AzureGroupID.String())
			if err != nil {
				return err
			}
			for _, gimpMember := range members {
				if !strings.HasSuffix(gimpMember.Email, cfg.TenantDomain) {
					log.Warnf("Skip member %q for team %q", gimpMember.Email, yamlteam.Name)
					continue
				}

				// check if member is already owner, if so, no need to add it to the member list as well
				if _, isOwner := teamOwners[gimpMember.Email]; isOwner {
					continue
				}

				if _, exists := users[gimpMember.Email]; !exists {
					member, err := dbtx.GetUserByEmail(ctx, gimpMember.Email)
					if err != nil {
						member, err = dbtx.CreateUser(ctx, gimpMember.Name, gimpMember.Email, gimpMember.ExternalID)
						if err != nil {
							return err
						}

						targets := []auditlogger.Target{
							auditlogger.UserTarget(member.Email),
						}
						fields := auditlogger.Fields{
							Action:        sqlc.AuditActionLegacyImporterUserCreate,
							CorrelationID: correlationID,
						}
						err = auditLogger.Logf(ctx, targets, fields, "Imported user from Azure AD")
						if err != nil {
							return err
						}
					}
					users[gimpMember.Email] = member
				}
				teamMembers[gimpMember.Email] = users[gimpMember.Email]
			}

			if len(teamOwners) == 0 && len(teamMembers) == 0 {
				log.Warnf("The Azure Group %q has no members or administrators.", yamlteam.Name)
			}

			if len(teamOwners) == 0 && len(teamMembers) > 0 {
				_, user := first(teamMembers)
				log.Infof("The Azure Group %q has no administrators, setting the first member as owner for the Console team: %q", yamlteam.Name, user.Email)
				teamOwners[user.Email] = user
				delete(teamMembers, user.Email)
			}

			convertedTeam, metadata := yamlteam.Convert()

			team, err := dbtx.GetTeamBySlug(ctx, convertedTeam.Slug)
			if err != nil {
				team, err = dbtx.CreateTeam(ctx, convertedTeam.Slug, convertedTeam.Purpose)
				if err != nil {
					return err
				}

				err = dbtx.SetTeamMetadata(ctx, team.Slug, metadata)
				if err != nil {
					return err
				}

				targets := []auditlogger.Target{
					auditlogger.TeamTarget(team.Slug),
				}
				fields := auditlogger.Fields{
					Action:        sqlc.AuditActionLegacyImporterTeamCreate,
					CorrelationID: correlationID,
				}
				err = auditLogger.Logf(ctx, targets, fields, "Imported team from Azure AD")
				if err != nil {
					return err
				}
			}

			for _, user := range teamOwners {
				err := dbtx.SetTeamMemberRole(ctx, user.ID, team.Slug, sqlc.RoleNameTeamowner)
				if err != nil {
					return err
				}

				targets := []auditlogger.Target{
					auditlogger.TeamTarget(team.Slug),
					auditlogger.UserTarget(user.Email),
				}
				fields := auditlogger.Fields{
					Action:        sqlc.AuditActionLegacyImporterTeamAddOwner,
					CorrelationID: correlationID,
				}
				err = auditLogger.Logf(ctx, targets, fields, "Assign %s as team owner from Azure AD", user.Email)
				if err != nil {
					return err
				}
			}

			for _, user := range teamMembers {
				err := dbtx.SetTeamMemberRole(ctx, user.ID, team.Slug, sqlc.RoleNameTeammember)
				if err != nil {
					return err
				}

				targets := []auditlogger.Target{
					auditlogger.TeamTarget(team.Slug),
					auditlogger.UserTarget(user.Email),
				}
				fields := auditlogger.Fields{
					Action:        sqlc.AuditActionLegacyImporterTeamAddMember,
					CorrelationID: correlationID,
				}
				err = auditLogger.Logf(ctx, targets, fields, "Assign %s as team member from Azure AD", user.Email)
				if err != nil {
					return err
				}
			}

			err = dbtx.SetReconcilerStateForTeam(ctx, sqlc.ReconcilerNameAzureGroup, team.Slug, reconcilers.AzureState{
				GroupID: &yamlteam.AzureGroupID,
			})
			if err != nil {
				return err
			}

			err = dbtx.SetReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGithubTeam, team.Slug, reconcilers.GitHubState{
				Slug: &team.Slug,
			})
			if err != nil {
				return err
			}

			googleWorkspaceGroupEmail := string(team.Slug) + "@" + cfg.TenantDomain
			err = dbtx.SetReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGoogleWorkspaceAdmin, team.Slug, reconcilers.GoogleWorkspaceState{
				GroupEmail: &googleWorkspaceGroupEmail,
			})
			if err != nil {
				return err
			}

			// Set GCP project state
			cachedMapping := gcpCache[string(team.Slug)]
			if len(cachedMapping) != len(clusters) {
				return fmt.Errorf("gcp team cache for %s has wrong size %d", team.Slug, len(cachedMapping))
			}
			projectMapping := make(map[string]reconcilers.GoogleGcpEnvironmentProject)
			for env := range clusters {
				projectMapping[env] = reconcilers.GoogleGcpEnvironmentProject{ProjectID: cachedMapping[env]}
			}
			err = dbtx.SetReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGoogleGcpProject, team.Slug, reconcilers.GoogleGcpProjectState{
				Projects: projectMapping,
			})
			if err != nil {
				return err
			}

			naisNamespaces := make(map[string]slug.Slug)
			for env := range clusters {
				naisNamespaces[env] = team.Slug
			}
			err = dbtx.SetReconcilerStateForTeam(ctx, sqlc.ReconcilerNameNaisNamespace, team.Slug, reconcilers.GoogleGcpNaisNamespaceState{
				Namespaces: naisNamespaces,
			})
			if err != nil {
				return err
			}

			_, err = dbtx.DisableTeam(ctx, team.Slug)
			if err != nil {
				return err
			}

			log.Infof("Created team %q with %d owners and %d members", team.Slug, len(teamOwners), len(teamMembers))
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	log.Infof("Done")

	return nil
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
