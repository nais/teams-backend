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
	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/apierror"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

// SetGitHubTeamSlug is the resolver for the setGitHubTeamSlug field.
func (r *mutationResolver) SetGitHubTeamSlug(ctx context.Context, teamSlug *slug.Slug, gitHubTeamSlug *slug.Slug) (*db.Team, error) {
	team, err := r.database.GetTeamBySlug(ctx, *teamSlug)
	if err != nil {
		return nil, apierror.ErrTeamNotExist
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	err = r.database.SetReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGithubTeam, *teamSlug, reconcilers.GitHubState{
		Slug: gitHubTeamSlug,
	})
	if err != nil {
		return nil, apierror.Errorf("Unable to save the GitHub state.")
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiReconcilersUpdateTeamState,
		Actor:         authz.ActorFromContext(ctx),
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Update GitHub state, set team slug: %q", gitHubTeamSlug)

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// SetGoogleWorkspaceGroupEmail is the resolver for the setGoogleWorkspaceGroupEmail field.
func (r *mutationResolver) SetGoogleWorkspaceGroupEmail(ctx context.Context, teamSlug *slug.Slug, googleWorkspaceGroupEmail string) (*db.Team, error) {
	team, err := r.database.GetTeamBySlug(ctx, *teamSlug)
	if err != nil {
		return nil, apierror.ErrTeamNotExist
	}

	if !strings.HasSuffix(googleWorkspaceGroupEmail, "@"+r.tenantDomain) {
		return nil, apierror.Errorf("Incorrect domain in email address %q. The required domain is %q.", googleWorkspaceGroupEmail, r.tenantDomain)
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	err = r.database.SetReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGoogleWorkspaceAdmin, *teamSlug, reconcilers.GoogleWorkspaceState{
		GroupEmail: &googleWorkspaceGroupEmail,
	})
	if err != nil {
		return nil, apierror.Errorf("Unable to save the Google Workspace state.")
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiReconcilersUpdateTeamState,
		Actor:         authz.ActorFromContext(ctx),
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Update Google Workspace state, set group email: %q", googleWorkspaceGroupEmail)

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// SetAzureADGroupID is the resolver for the setAzureADGroupId field.
func (r *mutationResolver) SetAzureADGroupID(ctx context.Context, teamSlug *slug.Slug, azureADGroupID *uuid.UUID) (*db.Team, error) {
	team, err := r.database.GetTeamBySlug(ctx, *teamSlug)
	if err != nil {
		return nil, apierror.ErrTeamNotExist
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	err = r.database.SetReconcilerStateForTeam(ctx, sqlc.ReconcilerNameAzureGroup, *teamSlug, reconcilers.AzureState{
		GroupID: azureADGroupID,
	})
	if err != nil {
		return nil, apierror.Errorf("Unable to save the Azure AD state.")
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiReconcilersUpdateTeamState,
		Actor:         authz.ActorFromContext(ctx),
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Update Azure AD state, set group ID: %q", azureADGroupID)

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// SetGcpProjectID is the resolver for the setGcpProjectId field.
func (r *mutationResolver) SetGcpProjectID(ctx context.Context, teamSlug *slug.Slug, gcpEnvironment string, gcpProjectID string) (*db.Team, error) {
	if len(r.gcpEnvironments) == 0 {
		return nil, apierror.Errorf("GCP cluster info has not been configured.")
	}

	team, err := r.database.GetTeamBySlug(ctx, *teamSlug)
	if err != nil {
		return nil, apierror.ErrTeamNotExist
	}

	if !console.Contains(r.gcpEnvironments, gcpEnvironment) {
		return nil, apierror.Errorf("Unknown GCP environment %q. Supported environments are: %s", gcpEnvironment, strings.Join(r.gcpEnvironments, ", "))
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	state := &reconcilers.GoogleGcpProjectState{
		Projects: make(map[string]reconcilers.GoogleGcpEnvironmentProject),
	}
	err = r.database.LoadReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGoogleGcpProject, *teamSlug, state)
	if err != nil {
		return nil, apierror.Errorf("Unable to load the existing GCP project state.")
	}

	state.Projects[gcpEnvironment] = reconcilers.GoogleGcpEnvironmentProject{ProjectID: gcpProjectID}
	err = r.database.SetReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGoogleGcpProject, *teamSlug, state)
	if err != nil {
		return nil, apierror.Errorf("Unable to save the GCP project state.")
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiReconcilersUpdateTeamState,
		Actor:         authz.ActorFromContext(ctx),
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Update GCP project state, set project ID %q in environment %q", gcpProjectID, gcpEnvironment)

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// SetNaisNamespace is the resolver for the setNaisNamespace field.
func (r *mutationResolver) SetNaisNamespace(ctx context.Context, teamSlug *slug.Slug, gcpEnvironment string, naisNamespace *slug.Slug) (*db.Team, error) {
	if len(r.gcpEnvironments) == 0 {
		return nil, apierror.Errorf("GCP cluster info has not been configured.")
	}

	team, err := r.database.GetTeamBySlug(ctx, *teamSlug)
	if err != nil {
		return nil, apierror.ErrTeamNotExist
	}

	if !console.Contains(r.gcpEnvironments, gcpEnvironment) {
		return nil, apierror.Errorf("Unknown GCP environment %q. Supported environments are: %s", gcpEnvironment, strings.Join(r.gcpEnvironments, ", "))
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	state := &reconcilers.GoogleGcpNaisNamespaceState{
		Namespaces: make(map[string]slug.Slug),
	}
	err = r.database.LoadReconcilerStateForTeam(ctx, sqlc.ReconcilerNameNaisNamespace, *teamSlug, state)
	if err != nil {
		return nil, apierror.Errorf("Unable to load the existing NAIS namespace state.")
	}

	state.Namespaces[gcpEnvironment] = *naisNamespace
	err = r.database.SetReconcilerStateForTeam(ctx, sqlc.ReconcilerNameNaisNamespace, *teamSlug, state)
	if err != nil {
		return nil, apierror.Errorf("Unable to save the NAIS namespace state.")
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiReconcilersUpdateTeamState,
		Actor:         authz.ActorFromContext(ctx),
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Update NAIS namespace state, set namespace %q in environment %q", naisNamespace, gcpEnvironment)

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}
