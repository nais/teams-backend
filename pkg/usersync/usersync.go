package usersync

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	"golang.org/x/oauth2/jwt"
	"gorm.io/gorm"

	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"golang.org/x/oauth2/google"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

type userSynchronizer struct {
	logger auditlogger.Logger
	domain string
	config *jwt.Config
	db     *gorm.DB
}

const (
	Name                    = "google:workspace-admin"
	OpCreate                = "google:workspace-admin:create"
	OpAddMember             = "google:workspace-admin:add-member"
	OpAddMembers            = "google:workspace-admin:add-members"
	OpDeleteMember          = "google:workspace-admin:delete-member"
	OpDeleteMembers         = "google:workspace-admin:delete-members"
	OpAddToGKESecurityGroup = "google:workspace-admin:add-to-gke-security-group"
)

func New(logger auditlogger.Logger, db *gorm.DB, domain string, config *jwt.Config) *userSynchronizer {
	return &userSynchronizer{
		config: config,
		db:     db,
		domain: domain,
		logger: logger,
	}
}

func NewFromConfig(cfg *config.Config, db *gorm.DB, logger auditlogger.Logger) (*userSynchronizer, error) {
	if !cfg.UserSync.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	b, err := ioutil.ReadFile(cfg.UserSync.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("read google credentials file: %w", err)
	}

	cf, err := google.JWTConfigFromJSON(
		b,
		admin_directory_v1.AdminDirectoryUserReadonlyScope,
		admin_directory_v1.AdminDirectoryGroupScope,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize google credentials: %w", err)
	}

	cf.Subject = cfg.UserSync.DelegatedUser

	return New(logger, db, cfg.UserSync.Domain, cf), nil
}

func (s *userSynchronizer) FetchAll(ctx context.Context) error {
	client := s.config.Client(ctx)
	srv, err := admin_directory_v1.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("retrieve directory client: %w", err)
	}

	resp, err := srv.Users.List().Domain(s.domain).Do()
	if err != nil {
		return fmt.Errorf("list remote users: %w", err)
	}

	userIds := make(map[uuid.UUID]struct{})

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Loop over all remote users and create them locally.
		// Existing users are ignored.
		for _, remoteUser := range resp.Users {
			localUser := &dbmodels.User{}

			// If we don't shadow this variable, it will bug out in the next iteration, i.e. query for two e-mail addresses.
			// BUG?
			tx := tx.First(localUser, "email = ?", remoteUser.PrimaryEmail)

			if tx.Error != nil {
				localUser := &dbmodels.User{
					Email: strp(remoteUser.PrimaryEmail),
					Name:  strp(remoteUser.Name.FullName),
				}

				tx = tx.Create(localUser)
				if tx.Error == nil {
					s.logger.UserLogf(reconcilers.Input{System: &dbmodels.System{}}, "usersync:create", localUser, "Local user created")
				}
			}

			if tx.Error != nil {
				return fmt.Errorf("create local user %s: %w", remoteUser.PrimaryEmail, tx.Error)
			}

			userIds[*localUser.ID] = struct{}{}
		}

		// Delete all local users with e-mail addresses that are not a part of the directory.
		localUsers := make([]*dbmodels.User, 0)
		domainEmails := "%@" + s.domain
		tx = tx.Find(&localUsers, "email LIKE ?", domainEmails)
		if tx.Error != nil {
			return fmt.Errorf("get local users: %w", tx.Error)
		}

		for _, localUser := range localUsers {
			_, touched := userIds[*localUser.ID]
			if touched {
				continue
			}

			tx = tx.Delete(localUser)
			if tx.Error != nil {
				return fmt.Errorf("delete local user %s: %w", localUser.Email, tx.Error)
			}
		}

		return tx.Error
	})

	return err
}

func strp(s string) *string {
	return &s
}
