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
	"google.golang.org/api/option"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"strings"

	"github.com/nais/console/pkg/config"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
)

type userSynchronizer struct {
	system      dbmodels.System
	auditLogger auditlogger.AuditLogger
	domain      string
	db          *gorm.DB
	client      *http.Client
}

const (
	OpPrepare    = "usersync:prepare"
	OpListRemote = "usersync:list:remote"
	OpListLocal  = "usersync:list:local"
	OpCreate     = "usersync:create"
	OpUpdate     = "usersync:update"
	OpDelete     = "usersync:delete"
)

var (
	ErrNotEnabled = errors.New("disabled by configuration")
)

func New(db *gorm.DB, system dbmodels.System, auditLogger auditlogger.AuditLogger, domain string, client *http.Client) *userSynchronizer {
	return &userSynchronizer{
		db:          db,
		system:      system,
		auditLogger: auditLogger,
		domain:      domain,
		client:      client,
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

	return New(db, system, auditLogger, cfg.TenantDomain, cf.Client(context.Background())), nil
}

type auditLogEntry struct {
	action  string
	user    dbmodels.User
	message string
}

// Sync Fetch all users from the tenant and add them as local users in Console. If a user already exists in Console
// the local user will get the name potentially updated. After all users have been upserted, local users that matches
// the tenant domain that does not exist in the Google Directory will be removed.
func (s *userSynchronizer) Sync(ctx context.Context) error {
	srv, err := admin_directory_v1.NewService(ctx, option.WithHTTPClient(s.client))
	if err != nil {
		return fmt.Errorf("%s: retrieve directory client: %w", OpListRemote, err)
	}

	resp, err := srv.Users.List().Context(ctx).Domain(s.domain).Do()
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

		defaultRoleNames := []roles.Role{
			roles.RoleTeamCreator,
			roles.RoleTeamViewer,
			roles.RoleUserViewer,
			roles.RoleServiceAccountCreator,
		}
		defaultRoles := make([]dbmodels.Role, 0)
		err = tx.Where("name IN (?)", defaultRoleNames).Find(&defaultRoles).Error
		if err != nil {
			return fmt.Errorf("%s: find default roles: %w", OpPrepare, err)
		}

		for _, remoteUser := range resp.Users {
			email := strings.ToLower(remoteUser.PrimaryEmail)
			localUser := &dbmodels.User{
				Email: email,
				Name:  remoteUser.Name.FullName,
			}

			ret := tx.Where("email = ?", email).FirstOrCreate(localUser)
			if ret.Error != nil {
				return fmt.Errorf("%s: create local user %s: %w", OpCreate, email, err)
			}

			if ret.RowsAffected > 0 {
				auditLogEntries = append(auditLogEntries, &auditLogEntry{
					action:  OpCreate,
					message: fmt.Sprintf("Local user created: %s", email),
					user:    *localUser,
				})
			}

			if localUser.Name != remoteUser.Name.FullName {
				localUser.Name = remoteUser.Name.FullName
				err = tx.Save(localUser).Error
				if err != nil {
					return fmt.Errorf("%s: update local user %s: %w", OpUpdate, email, err)
				}
				auditLogEntries = append(auditLogEntries, &auditLogEntry{
					action:  OpUpdate,
					message: fmt.Sprintf("Local user updated: %s", email),
					user:    *localUser,
				})
			}

			for _, role := range defaultRoles {
				err = tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "user_id"}, {Name: "role_id"}},
					DoNothing: true,
				}).Create(&dbmodels.UserRole{RoleID: *role.ID, UserID: *localUser.ID}).Error
				if err != nil {
					return fmt.Errorf("%s: attach default roles to user %s: %w", OpUpdate, email, err)
				}
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
			if _, upserted := userIds[*localUser.ID]; upserted {
				continue
			}

			err = tx.Delete(localUser).Error
			if err != nil {
				return fmt.Errorf("%s: delete local user %s: %w", OpDelete, localUser.Email, err)
			}

			auditLogEntries = append(auditLogEntries, &auditLogEntry{
				action:  OpDelete,
				message: fmt.Sprintf("Local user deleted: %s", localUser.Email),
				user:    *localUser,
			})
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, entry := range auditLogEntries {
		s.auditLogger.Logf(entry.action, *corr, s.system, nil, nil, &entry.user, entry.message)
	}

	return nil
}
