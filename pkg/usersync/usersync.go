package usersync

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/google_jwt"
	"github.com/nais/console/pkg/sqlc"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
	"net/http"
	"strings"
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
		return fmt.Errorf("%s: retrieve directory client: %w", sqlc.AuditActionUsersyncListRemote, err)
	}

	resp, err := srv.Users.List().Context(ctx).Domain(s.tenantDomain).Do()
	if err != nil {
		return fmt.Errorf("%s: list remote users: %w", sqlc.AuditActionUsersyncListRemote, err)
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("%s: unable to create UUID for correlation: %w", sqlc.AuditActionUsersyncPrepare, err)
	}

	auditLogEntries := make([]*auditLogEntry, 0)

	err = s.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		userIds := make(map[uuid.UUID]struct{})

		for _, remoteUser := range resp.Users {
			email := strings.ToLower(remoteUser.PrimaryEmail)
			localUser, err := dbtx.GetUserByEmail(ctx, email)
			if err != nil {
				localUser, err = dbtx.AddUser(ctx, remoteUser.Name.FullName, email)
				if err != nil {
					return fmt.Errorf("%s: create local user %s: %w", sqlc.AuditActionUsersyncCreate, email, err)
				}
				auditLogEntries = append(auditLogEntries, &auditLogEntry{
					action:    sqlc.AuditActionUsersyncCreate,
					message:   fmt.Sprintf("Local user created: %s", email),
					userEmail: localUser.Email,
				})
			}

			if localUser.Name != remoteUser.Name.FullName {
				localUser, err = dbtx.SetUserName(ctx, localUser.ID, remoteUser.Name.FullName)
				if err != nil {
					return fmt.Errorf("%s: update local user %s: %w", sqlc.AuditActionUsersyncUpdate, email, err)
				}
				auditLogEntries = append(auditLogEntries, &auditLogEntry{
					action:    sqlc.AuditActionUsersyncUpdate,
					message:   fmt.Sprintf("Local user updated: %s", email),
					userEmail: localUser.Email,
				})
			}

			for _, roleName := range DefaultRoleNames {
				err = dbtx.AssignGlobalRoleToUser(ctx, localUser.ID, roleName)
				if err != nil {
					return fmt.Errorf("%s: attach default role '%s' to user %s: %w", sqlc.AuditActionUsersyncUpdate, roleName, email, err)
				}
			}

			userIds[localUser.ID] = struct{}{}
		}

		localUsers, err := dbtx.GetUsersByEmail(ctx, "%@"+s.tenantDomain)
		if err != nil {
			return fmt.Errorf("%s: list local users: %w", sqlc.AuditActionUsersyncListLocal, err)
		}

		for _, localUser := range localUsers {
			if _, upserted := userIds[localUser.ID]; upserted {
				continue
			}

			err = dbtx.DeleteUser(ctx, localUser.ID)
			if err != nil {
				return fmt.Errorf("%s: delete local user %s: %w", sqlc.AuditActionUsersyncDelete, localUser.Email, err)
			}

			auditLogEntries = append(auditLogEntries, &auditLogEntry{
				action:    sqlc.AuditActionUsersyncDelete,
				message:   fmt.Sprintf("Local user deleted: %s", localUser.Email),
				userEmail: localUser.Email,
			})
		}

		return nil
	})

	if err != nil {
		return err
	}

	systemName := sqlc.SystemNameConsole

	for _, entry := range auditLogEntries {
		s.auditLogger.Logf(ctx, entry.action, correlationID, &systemName, nil, nil, &entry.userEmail, entry.message)
	}

	return nil
}
