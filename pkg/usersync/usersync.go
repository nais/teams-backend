package usersync

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/google_jwt"
	"github.com/nais/console/pkg/roles"
	"golang.org/x/oauth2/jwt"
	"gorm.io/gorm"

	"github.com/nais/console/pkg/config"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

type userSynchronizer struct {
	system      dbmodels.System
	auditLogger auditlogger.AuditLogger
	domain      string
	config      *jwt.Config
	db          *gorm.DB
}

const (
	OpPrepare    = "usersync:prepare"
	OpListRemote = "usersync:list:remote"
	OpListLocal  = "usersync:list:local"
	OpCreate     = "usersync:create"
	OpDelete     = "usersync:delete"
)

var (
	ErrNotEnabled = errors.New("disabled by configuration")
)

func New(db *gorm.DB, system dbmodels.System, auditLogger auditlogger.AuditLogger, domain string, config *jwt.Config) *userSynchronizer {
	return &userSynchronizer{
		db:          db,
		system:      system,
		auditLogger: auditLogger,
		domain:      domain,
		config:      config,
	}
}

func NewFromConfig(cfg *config.Config, db *gorm.DB, system dbmodels.System, auditLogger auditlogger.AuditLogger) (*userSynchronizer, error) {
	if !cfg.UserSync.Enabled {
		return nil, ErrNotEnabled
	}

	cf, err := google_jwt.GetConfig(cfg.Google.CredentialsFile, cfg.Google.DelegatedUser)

	if err != nil {
		return nil, fmt.Errorf("get google jwt config: %w", err)
	}

	return New(db, system, auditLogger, cfg.PartnerDomain, cf), nil
}

type auditLogEntry struct {
	action  string
	user    dbmodels.User
	message string
}

// Sync Fetch all users from the partner and add them as local users in Console. If a user already exists in Console
// the local user will remain untouched. After all users have been added we will also remove all local users that
// matches the partner domain that does not exist in the Google Directory.
// All new users will be grated two roles: "Team Creator" and "Team viewer"
func (s *userSynchronizer) Sync(ctx context.Context) error {
	defaultRoleIds, err := s.getDefaultRoleIds()
	if err != nil {
		return fmt.Errorf("%s: unable to fetch roles: %w", OpPrepare, err)
	}

	client := s.config.Client(ctx)
	srv, err := admin_directory_v1.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("%s: retrieve directory client: %w", OpListRemote, err)
	}

	resp, err := srv.Users.List().Domain(s.domain).Do()
	if err != nil {
		return fmt.Errorf("%s: list remote users: %w", OpListRemote, err)
	}

	corr := &dbmodels.Correlation{}
	err = s.db.Create(corr).Error
	if err != nil {
		return fmt.Errorf("%s: unable to create correlation for audit logs: %w", OpPrepare, err)
	}

	auditLogEntries := make([]*auditLogEntry, 0)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		userIds := make(map[uuid.UUID]struct{})

		for _, remoteUser := range resp.Users {
			localUser := &dbmodels.User{}

			err = tx.Where("email = ?", remoteUser.PrimaryEmail).First(localUser).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				localUser = &dbmodels.User{
					Email: remoteUser.PrimaryEmail,
					Name:  remoteUser.Name.FullName,
				}

				err = tx.Create(localUser).Error
				if err != nil {
					return fmt.Errorf("%s: create local user %s: %w", OpCreate, remoteUser.PrimaryEmail, err)
				}

				roleBindings := make([]dbmodels.RoleBinding, len(defaultRoleIds))
				for idx, roleId := range defaultRoleIds {
					roleBindings[idx] = dbmodels.RoleBinding{
						RoleID: roleId,
						UserID: *localUser.ID,
					}
				}

				err = tx.Create(roleBindings).Error
				if err != nil {
					return fmt.Errorf("%s: create role bindings for local user %s: %w", OpCreate, remoteUser.PrimaryEmail, err)
				}

				auditLogEntries = append(auditLogEntries, &auditLogEntry{
					action:  OpCreate,
					message: fmt.Sprintf("Local user created: %s", localUser.Name),
					user:    *localUser,
				})
			}

			userIds[*localUser.ID] = struct{}{}
		}

		localUsers := make([]*dbmodels.User, 0)
		domainEmails := "%@" + s.domain
		err = tx.Where("email LIKE ?", domainEmails).Find(&localUsers).Error
		if err != nil {
			return fmt.Errorf("%s: list local users: %w", OpListLocal, err)
		}

		for _, localUser := range localUsers {
			_, touched := userIds[*localUser.ID]
			if touched {
				continue
			}

			err = tx.Delete(localUser).Error
			if err != nil {
				return fmt.Errorf("%s: delete local user %s: %w", OpDelete, localUser.Email, err)
			}

			auditLogEntries = append(auditLogEntries, &auditLogEntry{
				action:  OpDelete,
				message: fmt.Sprintf("Local user deleted: %s", localUser.Name),
				user:    *localUser,
			})
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, entry := range auditLogEntries {
		s.auditLogger.Log(entry.action, *corr, s.system, nil, nil, &entry.user, entry.message)
	}

	return nil
}

func (s *userSynchronizer) getDefaultRoleIds() ([]uuid.UUID, error) {
	roles := []string{
		roles.TeamCreator,
		roles.TeamViewer,
		roles.RoleViewer,
	}
	roleIds := make([]uuid.UUID, len(roles))

	for idx, roleName := range roles {
		role := &dbmodels.Role{}
		err := s.db.Where("name = ?", roleName).First(role).Error
		if err != nil {
			return nil, fmt.Errorf("role not found %s: %w", roleName, err)
		}

		roleIds[idx] = *role.ID
	}

	return roleIds, nil
}
