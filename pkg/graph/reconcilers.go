package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/apierror"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/roles"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

// EnableReconciler is the resolver for the enableReconciler field.
func (r *mutationResolver) EnableReconciler(ctx context.Context, name sqlc.ReconcilerName) (*db.Reconciler, error) {
	if !name.Valid() {
		return nil, apierror.Errorf("%q is not a valid name", name)
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	_, err = r.database.GetReconciler(ctx, name)
	if err != nil {
		r.log.WithError(err).Errorf("unable to get reconciler: %q", name)
		return nil, apierror.Errorf("Unable to get reconciler: %q", name)
	}

	configs, err := r.database.GetReconcilerConfig(ctx, name)
	if err != nil {
		r.log.WithError(err).Errorf("unable to get reconciler config")
		return nil, apierror.Errorf("Unable to get reconciler config")
	}

	missingOptions := make([]string, 0)
	for _, config := range configs {
		if !config.Configured {
			missingOptions = append(missingOptions, string(config.Key))
		}
	}

	if len(missingOptions) != 0 {
		r.log.WithError(err).Errorf("reconciler is not fully configured")
		return nil, apierror.Errorf("Reconciler is not fully configured, missing one or more options: %s", strings.Join(missingOptions, ", "))
	}

	reconciler, err := r.database.EnableReconciler(ctx, name)
	if err != nil {
		r.log.WithError(err).Errorf("unable to enable reconciler")
		return nil, apierror.Errorf("Unable to enable reconciler")
	}

	err = r.teamSyncHandler.UseReconciler(*reconciler)
	if err != nil {
		if _, err := r.database.DisableReconciler(ctx, name); err != nil {
			r.log.WithError(err).Errorf("reconciler was enabled, but initialization failed, and we were unable to disable the reconciler.")
			return nil, apierror.Errorf("Reconciler was enabled, but initialization failed, and we were unable to disable the reconciler.")
		}

		r.log.WithError(err).Errorf("reconciler will not be enabled because of an initialization failure. Please verify that you have entered correct configuration values.")
		return nil, apierror.Errorf("Reconciler will not be enabled because of an initialization failure. Please verify that you have entered correct configuration values.")
	}

	actor := authz.ActorFromContext(ctx)
	targets := []auditlogger.Target{
		auditlogger.ReconcilerTarget(name),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiReconcilersEnable,
		Actor:         actor,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Enable reconciler: %q", name)

	_, err = r.teamSyncHandler.ScheduleAllTeams(ctx, correlationID)
	if err != nil {
		r.log.WithError(err).Errorf("reconcile all teams")
	}

	return reconciler, nil
}

// DisableReconciler is the resolver for the disableReconciler field.
func (r *mutationResolver) DisableReconciler(ctx context.Context, name sqlc.ReconcilerName) (*db.Reconciler, error) {
	if !name.Valid() {
		return nil, apierror.Errorf("%q is not a valid name", name)
	}

	var reconciler *db.Reconciler
	var err error
	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		reconciler, err = dbtx.GetReconciler(ctx, name)
		if err != nil {
			return err
		}

		reconciler, err = dbtx.DisableReconciler(ctx, name)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	actor := authz.ActorFromContext(ctx)
	targets := []auditlogger.Target{
		auditlogger.ReconcilerTarget(name),
	}
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiReconcilersDisable,
		Actor:  actor,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Disable reconciler: %q", name)
	r.teamSyncHandler.RemoveReconciler(name)

	return reconciler, nil
}

// ConfigureReconciler is the resolver for the configureReconciler field.
func (r *mutationResolver) ConfigureReconciler(ctx context.Context, name sqlc.ReconcilerName, config []*model.ReconcilerConfigInput) (*db.Reconciler, error) {
	if !name.Valid() {
		return nil, apierror.Errorf("%q is not a valid name", name)
	}

	reconcilerConfig := make(map[sqlc.ReconcilerConfigKey]string)
	for _, entry := range config {
		reconcilerConfig[entry.Key] = entry.Value
	}

	err := r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		rows, err := dbtx.GetReconcilerConfig(ctx, name)
		if err != nil {
			return err
		}

		validOptions := make(map[sqlc.ReconcilerConfigKey]struct{})
		for _, row := range rows {
			validOptions[row.Key] = struct{}{}
		}

		for key, value := range reconcilerConfig {
			if _, exists := validOptions[key]; !exists {
				keys := make([]string, 0, len(validOptions))
				for key := range validOptions {
					keys = append(keys, string(key))
				}
				return fmt.Errorf("unknown configuration option %q for reconciler %q. Valid options: %s", key, name, strings.Join(keys, ", "))
			}

			err = dbtx.ConfigureReconciler(ctx, name, key, value)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	reconciler, err := r.database.GetReconciler(ctx, name)
	if err != nil {
		return nil, err
	}

	if reconciler.Enabled {
		err = r.teamSyncHandler.UseReconciler(*reconciler)
		if err != nil {
			r.log.WithError(err).Errorf("use reconciler: %q", reconciler.Name)
		}
	}

	actor := authz.ActorFromContext(ctx)
	targets := []auditlogger.Target{
		auditlogger.ReconcilerTarget(name),
	}
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiReconcilersConfigure,
		Actor:  actor,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Configure reconciler: %q", name)

	return reconciler, nil
}

// ResetReconciler is the resolver for the resetReconciler field.
func (r *mutationResolver) ResetReconciler(ctx context.Context, name sqlc.ReconcilerName) (*db.Reconciler, error) {
	if !name.Valid() {
		return nil, fmt.Errorf("%q is not a valid name", name)
	}

	var reconciler *db.Reconciler
	var err error
	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		reconciler, err = dbtx.ResetReconcilerConfig(ctx, name)
		if err != nil {
			return err
		}

		if !reconciler.Enabled {
			return nil
		}

		reconciler, err = dbtx.DisableReconciler(ctx, name)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	actor := authz.ActorFromContext(ctx)
	targets := []auditlogger.Target{
		auditlogger.ReconcilerTarget(name),
	}
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiReconcilersReset,
		Actor:  actor,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Reset reconciler: %q", name)

	return reconciler, nil
}

// AddReconcilerOptOut is the resolver for the addReconcilerOptOut field.
func (r *mutationResolver) AddReconcilerOptOut(ctx context.Context, teamSlug *slug.Slug, userID *uuid.UUID, reconciler sqlc.ReconcilerName) (*model.TeamMember, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsUpdate, *teamSlug)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamBySlug(ctx, *teamSlug)
	if err != nil {
		return nil, apierror.ErrTeamNotExist
	}

	user, err := r.database.GetTeamMember(ctx, *teamSlug, *userID)
	if err != nil {
		return nil, apierror.ErrUserIsNotTeamMember
	}

	isOwner, err := r.database.UserIsTeamOwner(ctx, user.ID, *teamSlug)
	if err != nil {
		return nil, err
	}

	err = r.database.AddReconcilerOptOut(ctx, userID, teamSlug, reconciler)
	if err != nil {
		return nil, err
	}

	reconcilerOptOuts, err := r.database.GetTeamMemberOptOuts(ctx, user.ID, *teamSlug)
	if err != nil {
		return nil, err
	}

	role := model.TeamRoleMember
	if isOwner {
		role = model.TeamRoleOwner
	}

	return &model.TeamMember{
		Team:        team,
		User:        user,
		Role:        role,
		Reconcilers: reconcilerOptOuts,
	}, nil
}

// RemoveReconcilerOptOut is the resolver for the removeReconcilerOptOut field.
func (r *mutationResolver) RemoveReconcilerOptOut(ctx context.Context, teamSlug *slug.Slug, userID *uuid.UUID, reconciler sqlc.ReconcilerName) (*model.TeamMember, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsUpdate, *teamSlug)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamBySlug(ctx, *teamSlug)
	if err != nil {
		return nil, apierror.ErrTeamNotExist
	}

	user, err := r.database.GetTeamMember(ctx, *teamSlug, *userID)
	if err != nil {
		return nil, apierror.ErrUserIsNotTeamMember
	}

	isOwner, err := r.database.UserIsTeamOwner(ctx, user.ID, *teamSlug)
	if err != nil {
		return nil, err
	}

	err = r.database.RemoveReconcilerOptOut(ctx, userID, teamSlug, reconciler)
	if err != nil {
		return nil, err
	}

	reconcilerOptOuts, err := r.database.GetTeamMemberOptOuts(ctx, user.ID, *teamSlug)
	if err != nil {
		return nil, err
	}

	role := model.TeamRoleMember
	if isOwner {
		role = model.TeamRoleOwner
	}

	return &model.TeamMember{
		Team:        team,
		User:        user,
		Role:        role,
		Reconcilers: reconcilerOptOuts,
	}, nil
}

// Reconcilers is the resolver for the reconcilers field.
func (r *queryResolver) Reconcilers(ctx context.Context) ([]*db.Reconciler, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationTeamsCreate)
	if err != nil {
		return nil, err
	}

	return r.database.GetReconcilers(ctx)
}

// UsesTeamMemberships is the resolver for the usesTeamMemberships field.
func (r *reconcilerResolver) UsesTeamMemberships(ctx context.Context, obj *db.Reconciler) (bool, error) {
	switch obj.Name {
	case sqlc.ReconcilerNameGithubTeam:
		return true, nil
	case sqlc.ReconcilerNameAzureGroup:
		return true, nil
	case sqlc.ReconcilerNameGoogleWorkspaceAdmin:
		return true, nil
	case sqlc.ReconcilerNameNaisDependencytrack:
		return true, nil
	default:
		return false, nil
	}
}

// Config is the resolver for the config field.
func (r *reconcilerResolver) Config(ctx context.Context, obj *db.Reconciler) ([]*db.ReconcilerConfig, error) {
	return r.database.GetReconcilerConfig(ctx, obj.Name)
}

// Configured is the resolver for the configured field.
func (r *reconcilerResolver) Configured(ctx context.Context, obj *db.Reconciler) (bool, error) {
	configs, err := r.database.GetReconcilerConfig(ctx, obj.Name)
	if err != nil {
		return false, err
	}

	for _, config := range configs {
		if !config.Configured {
			return false, nil
		}
	}

	return true, nil
}

// AuditLogs is the resolver for the auditLogs field.
func (r *reconcilerResolver) AuditLogs(ctx context.Context, obj *db.Reconciler) ([]*db.AuditLog, error) {
	return r.database.GetAuditLogsForReconciler(ctx, obj.Name)
}

// Reconciler returns generated.ReconcilerResolver implementation.
func (r *Resolver) Reconciler() generated.ReconcilerResolver { return &reconcilerResolver{r} }

type reconcilerResolver struct{ *Resolver }
