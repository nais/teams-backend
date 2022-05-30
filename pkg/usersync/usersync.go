package usersync

import (
	"context"
	"errors"
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

	b, err := ioutil.ReadFile(cfg.Google.CredentialsFile)
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

	cf.Subject = cfg.Google.DelegatedUser

	return New(logger, db, cfg.Google.Domain, cf), nil
}

func (s *userSynchronizer) FetchAll(ctx context.Context) error {
	// dummy object for logging
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

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Loop over all remote users and create them locally.
		// Existing users are ignored.
		for _, remoteUser := range resp.Users {
			localUser := &dbmodels.User{}

			stmt := s.db.First(localUser, "email = ?", remoteUser.PrimaryEmail)
			if stmt.Error != nil {
				localUser = &dbmodels.User{
					Email: strp(remoteUser.PrimaryEmail),
					Name:  strp(remoteUser.Name.FullName),
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

	return err
}

func strp(s string) *string {
	return &s
}
