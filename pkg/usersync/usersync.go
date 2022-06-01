package usersync

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/google_jwt"
	"golang.org/x/oauth2/jwt"
	"gorm.io/gorm"

	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
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
	OpListRemote = "usersync:list:remote"
	OpListLocal  = "usersync:list:local"
	OpCreate     = "usersync:create"
	OpDelete     = "usersync:delete"
)

var (
	ErrNotEnabled = errors.New("disabled by configuration")
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
		return nil, ErrNotEnabled
	}

	cf, err := google_jwt.GetConfig(cfg.Google.CredentialsFile, cfg.Google.DelegatedUser)

	if err != nil {
		return nil, fmt.Errorf("get google jwt config: %w", err)
	}

	return New(logger, db, cfg.PartnerDomain, cf), nil
}

func (s *userSynchronizer) Sync(ctx context.Context) error {
	// dummy object for logging
	// FIXME: We should probably have a system for the user sync? Use the main console system for this perhaps?
	in := reconcilers.Input{System: &dbmodels.System{}}

	client := s.config.Client(ctx)
	srv, err := admin_directory_v1.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("retrieve directory client: %w", err)
	}

	resp, err := srv.Users.List().Domain(s.domain).Do()
	if err != nil {
		return s.logger.Errorf(in, OpListRemote, "list remote users: %w", err)
	}

	userIds := make(map[uuid.UUID]struct{})

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Loop over all remote users and create them locally.
		// Existing users are ignored.
		for _, remoteUser := range resp.Users {
			localUser := &dbmodels.User{}

			stmt := s.db.First(localUser, "email = ?", remoteUser.PrimaryEmail)
			if stmt.Error != nil {
				localUser = &dbmodels.User{
					Email: helpers.Strp(remoteUser.PrimaryEmail),
					Name:  helpers.Strp(remoteUser.Name.FullName),
				}

				tx = tx.Create(localUser)
				if tx.Error != nil {
					return s.logger.Errorf(in, OpCreate, "create local user %s: %w", remoteUser.PrimaryEmail, tx.Error)
				}

				s.logger.UserLogf(in, OpCreate, localUser, "Local user created")
			}

			userIds[*localUser.ID] = struct{}{}
		}

		// Delete all local users with e-mail addresses that are not a part of the directory.
		localUsers := make([]*dbmodels.User, 0)
		domainEmails := "%@" + s.domain
		tx = tx.Find(&localUsers, "email LIKE ?", domainEmails)
		if tx.Error != nil {
			return s.logger.Errorf(in, OpListLocal, "list local users: %w", tx.Error)
		}

		for _, localUser := range localUsers {
			_, touched := userIds[*localUser.ID]
			if touched {
				continue
			}

			tx = tx.Delete(localUser)
			if tx.Error != nil {
				return s.logger.Errorf(in, OpDelete, "delete local user %s: %w", *localUser.Email, tx.Error)
			}

			s.logger.UserLogf(in, OpDelete, localUser, "Local user deleted")
		}

		return tx.Error
	})
}
