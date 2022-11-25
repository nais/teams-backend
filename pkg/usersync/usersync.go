package usersync

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/google_token_source"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/sqlc"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type userSynchronizer struct {
	database     db.Database
	auditLogger  auditlogger.AuditLogger
	tenantDomain string
	service      *admin_directory_v1.Service
	log          logger.Logger
}

const adminGroupPrefix = "console-admins"

var (
	ErrNotEnabled    = errors.New("disabled by configuration")
	DefaultRoleNames = []sqlc.RoleName{
		sqlc.RoleNameTeamcreator,
		sqlc.RoleNameTeamviewer,
		sqlc.RoleNameUserviewer,
		sqlc.RoleNameServiceaccountcreator,
	}
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, domain string, service *admin_directory_v1.Service, log logger.Logger) *userSynchronizer {
	return &userSynchronizer{
		database:     database,
		auditLogger:  auditLogger,
		tenantDomain: domain,
		service:      service,
		log:          log,
	}
}

func NewFromConfig(cfg *config.Config, database db.Database, auditLogger auditlogger.AuditLogger, log logger.Logger) (*userSynchronizer, error) {
	log = log.WithSystem(string(sqlc.SystemNameUsersync))

	if !cfg.UserSync.Enabled {
		return nil, ErrNotEnabled
	}

	ctx := context.Background()

	ts, err := google_token_source.NewFromConfig(cfg).Admin(ctx)
	if err != nil {
		return nil, fmt.Errorf("create token source: %w", err)
	}

	srv, err := admin_directory_v1.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("retrieve directory client: %w", err)
	}

	return New(database, auditLogger, cfg.TenantDomain, srv, log), nil
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
	remoteUsers, err := getAllPaginatedUsers(ctx, s.service.Users, s.tenantDomain)
	if err != nil {
		return fmt.Errorf("list remote users: %w", err)
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("unable to create UUID for correlation: %w", err)
	}

	auditLogEntries := make([]auditLogEntry, 0)
	err = s.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		localUserIDs := make(map[uuid.UUID]struct{})
		remoteUserMapping := make(map[string]*db.User)

		for _, remoteUser := range remoteUsers {
			email := strings.ToLower(remoteUser.PrimaryEmail)
			localUser, created, err := getOrCreateLocalUserFromRemoteUser(ctx, dbtx, remoteUser)
			if err != nil {
				return fmt.Errorf("get or create local user %q: %w", email, err)
			}

			if created {
				auditLogEntries = append(auditLogEntries, auditLogEntry{
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

				auditLogEntries = append(auditLogEntries, auditLogEntry{
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

			localUserIDs[localUser.ID] = struct{}{}
			remoteUserMapping[remoteUser.Id] = localUser
		}

		err = deleteUnknownUsers(ctx, dbtx, localUserIDs, &auditLogEntries)
		if err != nil {
			return err
		}

		err = assignConsoleAdmins(ctx, dbtx, s.service.Members, s.tenantDomain, remoteUserMapping, &auditLogEntries, s.log)
		if err != nil {
			return err
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

// deleteUnknownUsers Delete users from the Console database that does not exist in the Google Workspace
func deleteUnknownUsers(ctx context.Context, dbtx db.Database, upsertedUsers map[uuid.UUID]struct{}, auditLogEntries *[]auditLogEntry) error {
	localUsers, err := dbtx.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("list local users: %w", err)
	}

	for _, localUser := range localUsers {
		if _, upserted := upsertedUsers[localUser.ID]; upserted {
			continue
		}

		err = dbtx.DeleteUser(ctx, localUser.ID)
		if err != nil {
			return fmt.Errorf("delete local user %q: %w", localUser.Email, err)
		}
		*auditLogEntries = append(*auditLogEntries, auditLogEntry{
			action:    sqlc.AuditActionUsersyncDelete,
			message:   fmt.Sprintf("Local user deleted: %q", localUser.Email),
			userEmail: localUser.Email,
		})
	}

	return nil
}

// assignConsoleAdmins Assign the global admin role to users based on the console-admins group. Existing admins that is
// not present in the list of admins will get the admin role revoked.
func assignConsoleAdmins(ctx context.Context, dbtx db.Database, membersService *admin_directory_v1.MembersService, domain string, remoteUserMapping map[string]*db.User, auditLogEntries *[]auditLogEntry, log logger.Logger) error {
	admins, err := getAdminUsers(ctx, membersService, domain, remoteUserMapping, log)
	if err != nil {
		return err
	}

	existingConsoleAdmins, err := getExistingConsoleAdmins(ctx, dbtx)
	if err != nil {
		return err
	}

	for _, existingAdmin := range existingConsoleAdmins {
		if _, shouldBeAdmin := admins[existingAdmin.ID]; !shouldBeAdmin {
			err = dbtx.RevokeGlobalUserRole(ctx, existingAdmin.ID, sqlc.RoleNameAdmin)
			if err != nil {
				return err
			}

			*auditLogEntries = append(*auditLogEntries, auditLogEntry{
				action:    sqlc.AuditActionUsersyncRevokeAdminRole,
				message:   fmt.Sprintf("Revoke global admin role from user: %q", existingAdmin.Email),
				userEmail: existingAdmin.Email,
			})
		}
	}

	for _, admin := range admins {
		if _, isAlreadyAdmin := existingConsoleAdmins[admin.ID]; !isAlreadyAdmin {
			err = dbtx.AssignGlobalRoleToUser(ctx, admin.ID, sqlc.RoleNameAdmin)
			if err != nil {
				return err
			}

			*auditLogEntries = append(*auditLogEntries, auditLogEntry{
				action:    sqlc.AuditActionUsersyncAssignAdminRole,
				message:   fmt.Sprintf("Assign global admin role to user: %q", admin.Email),
				userEmail: admin.Email,
			})
		}
	}

	return nil
}

// getExistingConsoleAdmins Get all users with a globally assigned admin role
func getExistingConsoleAdmins(ctx context.Context, dbtx db.Database) (map[uuid.UUID]*db.User, error) {
	users, err := dbtx.GetUsersWithGloballyAssignedRole(ctx, sqlc.RoleNameAdmin)
	if err != nil {
		return nil, err
	}

	existingAdmins := make(map[uuid.UUID]*db.User)
	for _, user := range users {
		existingAdmins[user.ID] = user
	}
	return existingAdmins, nil
}

// getAdminUsers Get a list of admin users based on the Console admins group in the Google Workspace
func getAdminUsers(ctx context.Context, membersService *admin_directory_v1.MembersService, domain string, remoteUserMapping map[string]*db.User, log logger.Logger) (map[uuid.UUID]*db.User, error) {
	adminGroupKey := adminGroupPrefix + "@" + domain
	groupMembers := make([]*admin_directory_v1.Member, 0)
	callback := func(fragments *admin_directory_v1.Members) error {
		for _, member := range fragments.Members {
			if member.Type == "USER" {
				groupMembers = append(groupMembers, member)
			}
		}
		return nil
	}
	admins := make(map[uuid.UUID]*db.User)
	err := membersService.
		List(adminGroupKey).
		Context(ctx).
		Pages(ctx, callback)
	if err != nil {
		if googleError, ok := err.(*googleapi.Error); ok && googleError.Code == 404 {
			// Special case: When the group does not exist we want to remove all existing admins. The group might never
			// have been created by the tenant admins in the first place, or it might have been deleted. In any case, we
			// want to treat this case as if the group exists, and that it is empty, effectively removing all admins.
			log.Warnf("console admins group %q does not exist", adminGroupKey)
			return admins, nil
		}

		return nil, fmt.Errorf("list members in Console admins group: %w", err)
	}

	for _, member := range groupMembers {
		admin, exists := remoteUserMapping[member.Id]
		if !exists {
			return nil, fmt.Errorf("uknown remote user")
		}

		admins[admin.ID] = admin
	}

	return admins, nil
}

// localUserIsOutdated Check if a local user is outdated when compared to the remote user
func localUserIsOutdated(localUser *db.User, remoteUser *admin_directory_v1.User) bool {
	return localUser.Name != remoteUser.Name.FullName ||
		!strings.EqualFold(localUser.Email, remoteUser.PrimaryEmail) ||
		localUser.ExternalID != remoteUser.Id
}

// getOrCreateLocalUserFromRemoteUser Look up the local user table for a match for the remote user. If no match is
// found, create the user.
func getOrCreateLocalUserFromRemoteUser(ctx context.Context, dbtx db.Database, remoteUser *admin_directory_v1.User) (localUser *db.User, created bool, err error) {
	localUser, err = dbtx.GetUserByExternalID(ctx, remoteUser.Id)
	if err == nil {
		return localUser, false, nil
	}

	email := strings.ToLower(remoteUser.PrimaryEmail)
	localUser, err = dbtx.GetUserByEmail(ctx, email)
	if err == nil {
		return localUser, false, nil
	}

	localUser, err = dbtx.CreateUser(ctx, remoteUser.Name.FullName, email, remoteUser.Id)
	if err != nil {
		return nil, false, err
	}

	return localUser, true, nil
}

func getAllPaginatedUsers(ctx context.Context, svc *admin_directory_v1.UsersService, domain string) ([]*admin_directory_v1.User, error) {
	users := make([]*admin_directory_v1.User, 0)

	callback := func(fragments *admin_directory_v1.Users) error {
		users = append(users, fragments.Users...)
		return nil
	}

	err := svc.
		List().
		Context(ctx).
		Domain(domain).
		ShowDeleted("false").
		Query("isSuspended=false").
		Pages(ctx, callback)

	return users, err
}
