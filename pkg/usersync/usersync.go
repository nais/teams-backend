package usersync

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/teams-backend/pkg/types"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/google_token_source"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/sqlc"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type (
	UserSynchronizer struct {
		database         db.Database
		auditLogger      auditlogger.AuditLogger
		adminGroupPrefix string
		tenantDomain     string
		service          *admin_directory_v1.Service
		log              logger.Logger
		syncRuns         *RunsHandler
	}

	auditLogEntry struct {
		action    types.AuditAction
		userEmail string
		message   string
	}

	// Key is the ID from Azure AD
	remoteUsersMap map[string]*db.User

	userMap struct {
		// byExternalID key is the ID from Azure AD
		byExternalID map[string]*db.User
		byEmail      map[string]*db.User
	}

	userByIDMap  map[uuid.UUID]*db.User
	userRolesMap map[*db.User]map[sqlc.RoleName]struct{}
)

var DefaultRoleNames = []sqlc.RoleName{
	sqlc.RoleNameTeamcreator,
	sqlc.RoleNameTeamviewer,
	sqlc.RoleNameUserviewer,
	sqlc.RoleNameServiceaccountcreator,
}

func New(database db.Database, auditLogger auditlogger.AuditLogger, adminGroupPrefix, tenantDomain string, service *admin_directory_v1.Service, log logger.Logger, syncRuns *RunsHandler) *UserSynchronizer {
	return &UserSynchronizer{
		database:         database,
		auditLogger:      auditLogger,
		adminGroupPrefix: adminGroupPrefix,
		tenantDomain:     tenantDomain,
		service:          service,
		log:              log,
		syncRuns:         syncRuns,
	}
}

func NewFromConfig(cfg *config.Config, database db.Database, log logger.Logger, syncRuns *RunsHandler) (*UserSynchronizer, error) {
	log = log.WithComponent(types.ComponentNameUsersync)
	ctx := context.Background()

	builder, err := google_token_source.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	ts, err := builder.Admin(ctx)
	if err != nil {
		return nil, fmt.Errorf("create token source: %w", err)
	}

	srv, err := admin_directory_v1.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("retrieve directory client: %w", err)
	}

	return New(database, auditlogger.New(database, types.ComponentNameUsersync, log), cfg.UserSync.AdminGroupPrefix, cfg.TenantDomain, srv, log, syncRuns), nil
}

// Sync Fetch all users from the tenant and add them as local users in teams-backend. If a user already exists in
// teams-backend the local user will get the name potentially updated. After all users have been upserted, local users
// that matches the tenant domain that does not exist in the Google Directory will be removed.
func (s *UserSynchronizer) Sync(ctx context.Context, correlationID uuid.UUID) error {
	log := s.log.WithCorrelationID(correlationID)
	syncRun := s.syncRuns.StartNewRun(correlationID)
	defer syncRun.Finish()

	remoteUserMapping := make(remoteUsersMap)
	remoteUsers, err := getAllPaginatedUsers(ctx, s.service.Users, s.tenantDomain)
	if err != nil {
		syncRun.FinishWithError(err)
		return fmt.Errorf("get remote users: %w", err)
	}

	auditLogEntries := make([]auditLogEntry, 0)
	err = s.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		allUsersRows, err := dbtx.GetUsers(ctx)
		if err != nil {
			return fmt.Errorf("get existing users: %w", err)
		}

		usersByID := make(userByIDMap)
		existingUsers := userMap{
			byExternalID: make(map[string]*db.User),
			byEmail:      make(map[string]*db.User),
		}

		for _, user := range allUsersRows {
			usersByID[user.ID] = user
			existingUsers.byExternalID[user.ExternalID] = user
			existingUsers.byEmail[user.Email] = user
		}

		allUserRolesRows, err := dbtx.GetAllUserRoles(ctx)
		if err != nil {
			return fmt.Errorf("get existing user roles: %w", err)
		}

		userRoles := make(userRolesMap)
		for _, row := range allUserRolesRows {
			user := usersByID[row.UserID]
			if _, exists := userRoles[user]; !exists {
				userRoles[user] = make(map[sqlc.RoleName]struct{})
			}
			userRoles[user][row.RoleName] = struct{}{}
		}

		for _, remoteUser := range remoteUsers {
			email := strings.ToLower(remoteUser.PrimaryEmail)
			localUser, created, err := getOrCreateLocalUserFromRemoteUser(ctx, dbtx, remoteUser, existingUsers)
			if err != nil {
				return fmt.Errorf("get or create local user %q: %w", email, err)
			}

			if created {
				auditLogEntries = append(auditLogEntries, auditLogEntry{
					action:    types.AuditActionUsersyncCreate,
					message:   fmt.Sprintf("Local user created: %q, external ID: %q", localUser.Email, localUser.ExternalID),
					userEmail: localUser.Email,
				})
			}

			if localUserIsOutdated(localUser, remoteUser) {
				updatedUser, err := dbtx.UpdateUser(ctx, localUser.ID, remoteUser.Name.FullName, email, remoteUser.Id)
				if err != nil {
					return fmt.Errorf("update local user %q: %w", email, err)
				}

				auditLogEntries = append(auditLogEntries, auditLogEntry{
					action:    types.AuditActionUsersyncUpdate,
					message:   fmt.Sprintf("Local user updated: %q, external ID: %q", updatedUser.Email, updatedUser.ExternalID),
					userEmail: updatedUser.Email,
				})
			}

			for _, roleName := range DefaultRoleNames {
				if globalRoles, userHasGlobalRoles := userRoles[localUser]; userHasGlobalRoles {
					if _, userHasDefaultRole := globalRoles[roleName]; userHasDefaultRole {
						continue
					}
				}
				err = dbtx.AssignGlobalRoleToUser(ctx, localUser.ID, roleName)
				if err != nil {
					return fmt.Errorf("attach default role %q to user %q: %w", roleName, email, err)
				}
			}

			remoteUserMapping[remoteUser.Id] = localUser
			delete(usersByID, localUser.ID)
		}

		deletedUsers, err := deleteUnknownUsers(ctx, dbtx, usersByID, &auditLogEntries)
		if err != nil {
			return err
		}

		for _, deletedUser := range deletedUsers {
			delete(userRoles, deletedUser)
		}

		err = assignTeamsBackendAdmins(ctx, dbtx, s.service.Members, s.adminGroupPrefix, s.tenantDomain, remoteUserMapping, userRoles, &auditLogEntries, log)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		syncRun.FinishWithError(err)
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

// deleteUnknownUsers Delete users from the teams-backend database that does not exist in the Google Workspace
func deleteUnknownUsers(ctx context.Context, dbtx db.Database, unknownUsers userByIDMap, auditLogEntries *[]auditLogEntry) ([]*db.User, error) {
	deletedUsers := make([]*db.User, 0)
	for _, user := range unknownUsers {
		err := dbtx.DeleteUser(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("delete local user %q: %w", user.Email, err)
		}
		*auditLogEntries = append(*auditLogEntries, auditLogEntry{
			action:    types.AuditActionUsersyncDelete,
			message:   fmt.Sprintf("Local user deleted: %q, external ID: %q", user.Email, user.ExternalID),
			userEmail: user.Email,
		})
		deletedUsers = append(deletedUsers, user)
	}

	return deletedUsers, nil
}

// assignTeamsBackendAdmins Assign the global admin role to users based on the admin group. Existing admins that is not
// present in the list of admins will get the admin role revoked.
func assignTeamsBackendAdmins(ctx context.Context, dbtx db.Database, membersService *admin_directory_v1.MembersService, adminGroupPrefix, tenantDomain string, remoteUserMapping map[string]*db.User, userRoles userRolesMap, auditLogEntries *[]auditLogEntry, log logger.Logger) error {
	admins, err := getAdminUsers(ctx, membersService, adminGroupPrefix, tenantDomain, remoteUserMapping, log)
	if err != nil {
		return err
	}

	existingAdmins := getExistingTeamsBackendAdmins(userRoles)
	for _, existingAdmin := range existingAdmins {
		if _, shouldBeAdmin := admins[existingAdmin.ID]; !shouldBeAdmin {
			err = dbtx.RevokeGlobalUserRole(ctx, existingAdmin.ID, sqlc.RoleNameAdmin)
			if err != nil {
				return err
			}

			*auditLogEntries = append(*auditLogEntries, auditLogEntry{
				action:    types.AuditActionUsersyncRevokeAdminRole,
				message:   fmt.Sprintf("Revoke global admin role from user: %q", existingAdmin.Email),
				userEmail: existingAdmin.Email,
			})
		}
	}

	for _, admin := range admins {
		if _, isAlreadyAdmin := existingAdmins[admin.ID]; !isAlreadyAdmin {
			err = dbtx.AssignGlobalRoleToUser(ctx, admin.ID, sqlc.RoleNameAdmin)
			if err != nil {
				return err
			}

			*auditLogEntries = append(*auditLogEntries, auditLogEntry{
				action:    types.AuditActionUsersyncAssignAdminRole,
				message:   fmt.Sprintf("Assign global admin role to user: %q", admin.Email),
				userEmail: admin.Email,
			})
		}
	}

	return nil
}

// getExistingTeamsBackendAdmins Get all users with a globally assigned admin role
func getExistingTeamsBackendAdmins(userWithRoles userRolesMap) map[uuid.UUID]*db.User {
	admins := make(map[uuid.UUID]*db.User)
	for user, roles := range userWithRoles {
		for roleName := range roles {
			if roleName == sqlc.RoleNameAdmin {
				admins[user.ID] = user
			}
		}
	}
	return admins
}

// getAdminUsers Get a list of admin users based on the teams-backend admins group in the Google Workspace
func getAdminUsers(ctx context.Context, membersService *admin_directory_v1.MembersService, adminGroupPrefix, tenantDomain string, remoteUserMapping map[string]*db.User, log logger.Logger) (map[uuid.UUID]*db.User, error) {
	adminGroupKey := adminGroupPrefix + "@" + tenantDomain
	groupMembers := make([]*admin_directory_v1.Member, 0)
	callback := func(fragments *admin_directory_v1.Members) error {
		for _, member := range fragments.Members {
			if member.Type == "USER" && member.Status == "ACTIVE" {
				groupMembers = append(groupMembers, member)
			}
		}
		return nil
	}
	admins := make(map[uuid.UUID]*db.User)
	err := membersService.
		List(adminGroupKey).
		IncludeDerivedMembership(true).
		Context(ctx).
		Pages(ctx, callback)
	if err != nil {
		if googleError, ok := err.(*googleapi.Error); ok && googleError.Code == 404 {
			// Special case: When the group does not exist we want to remove all existing admins. The group might never
			// have been created by the tenant admins in the first place, or it might have been deleted. In any case, we
			// want to treat this case as if the group exists, and that it is empty, effectively removing all admins.
			log.Warnf("teams-backend admins group %q does not exist", adminGroupKey)
			return admins, nil
		}

		return nil, fmt.Errorf("list members in teams-backend admins group: %w", err)
	}

	for _, member := range groupMembers {
		admin, exists := remoteUserMapping[member.Id]
		if !exists {
			log.Errorf("unknown user %q in admins groups", member.Email)
			continue
		}

		admins[admin.ID] = admin
	}

	return admins, nil
}

// localUserIsOutdated Check if a local user is outdated when compared to the remote user
func localUserIsOutdated(localUser *db.User, remoteUser *admin_directory_v1.User) bool {
	if localUser.Name != remoteUser.Name.FullName {
		return true
	}

	if !strings.EqualFold(localUser.Email, remoteUser.PrimaryEmail) {
		return true
	}

	if localUser.ExternalID != remoteUser.Id {
		return true
	}

	return false
}

// getOrCreateLocalUserFromRemoteUser Look up the local user table for a match for the remote user. If no match is
// found, create the user.
func getOrCreateLocalUserFromRemoteUser(ctx context.Context, dbtx db.Database, remoteUser *admin_directory_v1.User, existingUsers userMap) (*db.User, bool, error) {
	if existingUser, exists := existingUsers.byExternalID[remoteUser.Id]; exists {
		return existingUser, false, nil
	}

	email := strings.ToLower(remoteUser.PrimaryEmail)
	if existingUser, exists := existingUsers.byEmail[email]; exists {
		return existingUser, false, nil
	}

	createdUser, err := dbtx.CreateUser(ctx, remoteUser.Name.FullName, email, remoteUser.Id)
	if err != nil {
		return nil, false, err
	}

	return createdUser, true, nil
}

func getAllPaginatedUsers(ctx context.Context, svc *admin_directory_v1.UsersService, tenantDomain string) ([]*admin_directory_v1.User, error) {
	users := make([]*admin_directory_v1.User, 0)

	callback := func(fragments *admin_directory_v1.Users) error {
		users = append(users, fragments.Users...)
		return nil
	}

	err := svc.
		List().
		Context(ctx).
		Domain(tenantDomain).
		ShowDeleted("false").
		Query("isSuspended=false").
		Pages(ctx, callback)

	return users, err
}
