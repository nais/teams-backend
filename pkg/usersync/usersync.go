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
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	"github.com/nais/console/pkg/roles"
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

// Sync Fetch all users from the partner and add them as local users in Console. If a user already exists in Console
// the local user will remain untouched. After all users have been added we will also remove all local users that
// matches the partner domain that does not exist in the Google Directory.
// All new users will be grated two roles: "Team Creator" and "Team viewer"
func (s *userSynchronizer) Sync(ctx context.Context) error {
	tx := s.db.WithContext(ctx)

	system := &dbmodels.System{}
	err := tx.Where("name = ?", console_reconciler.Name).First(system).Error
	if err != nil {
		return err
	}

	in := reconcilers.Input{System: system}

	teamCreator := &dbmodels.Role{}
	err = tx.Where("name = ?", roles.TeamCreator).First(teamCreator).Error
	if err != nil {
		return s.logger.Errorf(in, OpCreate, "role not found %s: %w", roles.TeamCreator, err)
	}

	teamViewer := &dbmodels.Role{}
	err = tx.Where("name = ?", roles.TeamViewer).First(teamViewer).Error
	if err != nil {
		return s.logger.Errorf(in, OpCreate, "role not found %s: %w", roles.TeamViewer, err)
	}

	client := s.config.Client(ctx)
	srv, err := admin_directory_v1.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return s.logger.Errorf(in, OpListRemote, "retrieve directory client: %w", err)
	}

	resp, err := srv.Users.List().Domain(s.domain).Do()
	if err != nil {
		return s.logger.Errorf(in, OpListRemote, "list remote users: %w", err)
	}

	userIds := make(map[uuid.UUID]struct{})

	return tx.Transaction(func(tx *gorm.DB) error {
		for _, remoteUser := range resp.Users {
			localUser := &dbmodels.User{}

			err = tx.Where("email = ?", remoteUser.PrimaryEmail).First(localUser).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				localUser = &dbmodels.User{
					Email: helpers.Strp(remoteUser.PrimaryEmail),
					Name:  helpers.Strp(remoteUser.Name.FullName),
					RoleBindings: []*dbmodels.RoleBinding{
						{
							RoleID: teamCreator.ID,
						},
						{
							RoleID: teamViewer.ID,
						},
					},
				}

				err = tx.Create(localUser).Error
				if err != nil {
					return s.logger.Errorf(in, OpCreate, "create local user %s: %w", remoteUser.PrimaryEmail, err)
				}

				s.logger.UserLogf(in, OpCreate, localUser, "Local user created")

			}

			userIds[*localUser.ID] = struct{}{}
		}

		localUsers := make([]*dbmodels.User, 0)
		domainEmails := "%@" + s.domain
		err = tx.Where("email LIKE ?", domainEmails).Find(&localUsers).Error
		if err != nil {
			return s.logger.Errorf(in, OpListLocal, "list local users: %w", err)
		}

		for _, localUser := range localUsers {
			_, touched := userIds[*localUser.ID]
			if touched {
				continue
			}

			err = tx.Delete(localUser).Error
			if err != nil {
				return s.logger.Errorf(in, OpDelete, "delete local user %s: %w", *localUser.Email, err)
			}

			s.logger.UserLogf(in, OpDelete, localUser, "Local user deleted")
		}

		return nil
	})
}
