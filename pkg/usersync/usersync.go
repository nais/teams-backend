package usersync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/google_jwt"
	"github.com/nais/console/pkg/sqlc"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

type userSynchronizer struct {
	database     db.Database
	auditLogger  auditlogger.AuditLogger
	tenantDomain string
	client       *http.Client
}

var (
	ErrNotEnabled    = errors.New("disabled by configuration")
	DefaultRoleNames = []sqlc.RoleName{
		sqlc.RoleNameTeamcreator,
		sqlc.RoleNameTeamviewer,
		sqlc.RoleNameUserviewer,
		sqlc.RoleNameServiceaccountcreator,
	}
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, domain string, client *http.Client) *userSynchronizer {
	return &userSynchronizer{
		database:     database,
		auditLogger:  auditLogger,
		tenantDomain: domain,
		client:       client,
	}
}

func NewFromConfig(cfg *config.Config, database db.Database, auditLogger auditlogger.AuditLogger) (*userSynchronizer, error) {
	if !cfg.UserSync.Enabled {
		return nil, ErrNotEnabled
	}

	cf, err := google_jwt.GetConfig(cfg.Google.CredentialsFile, cfg.Google.DelegatedUser)
	if err != nil {
		return nil, fmt.Errorf("get google jwt config: %w", err)
	}

	return New(database, auditLogger, cfg.TenantDomain, cf.Client(context.Background())), nil
}

type auditLogEntry struct {
	action    sqlc.AuditAction
	userEmail string
	message   string
}

// Sync Fetch all users from the tenant and add them as local users in Console. If a user already exists in Console
// the local user will get the name potentially updated. After all users have been upserted, local users that matches
// the tenant domain that does not exist in the Google Directory will be removed.
func (s *userSynchronizer) Sync(ctx context.Context) error {
	srv, err := admin_directory_v1.NewService(ctx, option.WithHTTPClient(s.client))
	if err != nil {
		return fmt.Errorf("retrieve directory client: %w", err)
	}

	resp, err := srv.Users.List().Context(ctx).Domain(s.tenantDomain).Do()
	if err != nil {
		return fmt.Errorf("list remote users: %w", err)
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("unable to create UUID for correlation: %w", err)
	}

	auditLogEntries := make([]*auditLogEntry, 0)

	err = s.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		userIDs := make(map[uuid.UUID]struct{})

		for _, remoteUser := range resp.Users {
			email := strings.ToLower(remoteUser.PrimaryEmail)
			localUser, created, err := getOrCreateLocalUserFromRemoteUser(ctx, dbtx, remoteUser)
			if err != nil {
				return fmt.Errorf("get or create local user %q: %w", email, err)
			}

			if created {
				auditLogEntries = append(auditLogEntries, &auditLogEntry{
					action:    sqlc.AuditActionUsersyncCreate,
					message:   fmt.Sprintf("Local user created: %q", localUser.Email),
					userEmail: localUser.Email,
				})
			}

			if localUserIsOutdated(localUser, remoteUser) {
				updatedUser, err := dbtx.UpdateUser(ctx, localUser.ID, remoteUser.Name.FullName, email, remoteUser.Id)
				if err != nil {
					return fmt.Errorf("update local user %q: %w", email, err)
				}

				auditLogEntries = append(auditLogEntries, &auditLogEntry{
					action:    sqlc.AuditActionUsersyncUpdate,
					message:   fmt.Sprintf("Local user updated: %q", updatedUser.Email),
					userEmail: updatedUser.Email,
				})
			}

			for _, roleName := range DefaultRoleNames {
				err = dbtx.AssignGlobalRoleToUser(ctx, localUser.ID, roleName)
				if err != nil {
					return fmt.Errorf("attach default role %q to user %q: %w", roleName, email, err)
				}
			}

			userIDs[localUser.ID] = struct{}{}
		}

		localUsers, err := dbtx.GetUsers(ctx)
		if err != nil {
			return fmt.Errorf("list local users: %w", err)
		}

		for _, localUser := range localUsers {
			if _, upserted := userIDs[localUser.ID]; upserted {
				continue
			}

			err = dbtx.DeleteUser(ctx, localUser.ID)
			if err != nil {
				return fmt.Errorf("delete local user %q: %w", localUser.Email, err)
			}

			auditLogEntries = append(auditLogEntries, &auditLogEntry{
				action:    sqlc.AuditActionUsersyncDelete,
				message:   fmt.Sprintf("Local user deleted: %q", localUser.Email),
				userEmail: localUser.Email,
			})
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, entry := range auditLogEntries {
		targets := []auditlogger.Target{
			auditlogger.UserTarget(entry.userEmail),
		}
		fields := auditlogger.Fields{
			Action:        entry.action,
			CorrelationID: correlationID,
		}
		s.auditLogger.Logf(ctx, targets, fields, entry.message)
	}

	return nil
}

// localUserIsOutdated Check if a local user is outdated when compared to the remote user
func localUserIsOutdated(localUser *db.User, remoteUser *admin_directory_v1.User) bool {
	return localUser.Name != remoteUser.Name.FullName ||
		localUser.Email != remoteUser.PrimaryEmail ||
		localUser.ExternalID != remoteUser.Id
}

// getOrCreateLocalUserFromRemoteUser Look up the local user table for a match for the remote user. If no match is
// found, create the user.
func getOrCreateLocalUserFromRemoteUser(ctx context.Context, db db.Database, remoteUser *admin_directory_v1.User) (localUser *db.User, created bool, err error) {
	localUser, err = db.GetUserByExternalID(ctx, remoteUser.Id)
	if err == nil {
		return localUser, false, nil
	}

	email := strings.ToLower(remoteUser.PrimaryEmail)
	localUser, err = db.GetUserByEmail(ctx, email)
	if err == nil {
		return localUser, false, nil
	}

	localUser, err = db.CreateUser(ctx, remoteUser.Name.FullName, email, remoteUser.Id)
	if err != nil {
		return nil, false, err
	}

	return localUser, true, nil
}
