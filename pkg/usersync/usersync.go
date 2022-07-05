package usersync

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/google_jwt"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"

	"github.com/nais/console/pkg/config"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
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
	OpUpsert     = "usersync:upsert"
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

	return New(db, system, auditLogger, cfg.TenantDomain, cf), nil
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
		// Map of user IDs that is upserted
		userIds := make(map[uuid.UUID]struct{})

		for _, remoteUser := range resp.Users {
			email := strings.ToLower(remoteUser.PrimaryEmail)
			localUser := &dbmodels.User{
				Email: email,
				Name:  remoteUser.Name.FullName,
			}

			err = tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "email"}},
				DoUpdates: clause.AssignmentColumns([]string{"name"}),
			}).Create(&localUser).Error
			if err != nil {
				return fmt.Errorf("%s: upsert local user %s: %w", OpUpsert, email, err)
			}

			auditLogEntries = append(auditLogEntries, &auditLogEntry{
				action:  OpUpsert,
				message: fmt.Sprintf("Local user upserted: %s", email),
				user:    *localUser,
			})

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
