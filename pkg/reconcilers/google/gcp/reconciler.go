package google_gcp_reconciler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/gcp"
	"github.com/nais/console/pkg/google_token_source"
	"github.com/nais/console/pkg/legacy/envmap"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/metrics"
	"github.com/nais/console/pkg/reconcilers"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/serviceusage/v1"
)

const (
	Name                              = sqlc.ReconcilerNameGoogleGcpProject
	GoogleProjectDisplayNameMaxLength = 30
	metricsSystemName                 = "gcp"
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, clusters gcp.Clusters, gcpServices *GcpServices, tenantName, domain, cnrmRoleName, billingAccount string, log logger.Logger, legacyMapping []envmap.EnvironmentMapping) *googleGcpReconciler {
	return &googleGcpReconciler{
		database:       database,
		auditLogger:    auditLogger.WithSystemName(sqlc.SystemNameGoogleGcpProject),
		clusters:       clusters,
		gcpServices:    gcpServices,
		domain:         domain,
		cnrmRoleName:   cnrmRoleName,
		billingAccount: billingAccount,
		tenantName:     tenantName,
		log:            log.WithSystem(string(Name)),
		legacyMapping:  legacyMapping,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	gcpServices, err := createGcpServices(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return New(database, auditLogger, cfg.GCP.Clusters, gcpServices, cfg.TenantName, cfg.TenantDomain, cfg.GCP.CnrmRole, cfg.GCP.BillingAccount, log, cfg.LegacyNaisNamespaces), nil
}

func (r *googleGcpReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *googleGcpReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.GoogleGcpProjectState{
		Projects: make(map[string]reconcilers.GoogleGcpEnvironmentProject),
	}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, state)
	if err != nil {
		return fmt.Errorf("load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	if state.Projects == nil {
		state.Projects = make(map[string]reconcilers.GoogleGcpEnvironmentProject)
	}

	googleWorkspaceState := &reconcilers.GoogleWorkspaceState{}
	err = r.database.LoadReconcilerStateForTeam(ctx, google_workspace_admin_reconciler.Name, input.Team.Slug, googleWorkspaceState)
	if err != nil {
		return fmt.Errorf("load system state for team %q in system %q: %w", input.Team.Slug, google_workspace_admin_reconciler.Name, err)
	}

	if googleWorkspaceState.GroupEmail == nil {
		return fmt.Errorf("no Google Workspace group exists for team %q yet, is the %q reconciler enabled? ", input.Team.Slug, google_workspace_admin_reconciler.Name)
	}

	teamProjects := make(map[string]*cloudresourcemanager.Project, len(r.clusters))
	for environment, cluster := range r.clusters {
		projectID := GenerateProjectID(r.domain, environment, input.Team.Slug)
		teamProject, err := r.getOrCreateProject(ctx, projectID, state, environment, cluster.TeamsFolderID, input)
		if err != nil {
			return fmt.Errorf("get or create a GCP project %q for team %q in environment %q: %w", projectID, input.Team.Slug, environment, err)
		}
		teamProjects[environment] = teamProject
		state.Projects[environment] = reconcilers.GoogleGcpEnvironmentProject{
			ProjectID: teamProject.ProjectId,
		}

		err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, state)
		if err != nil {
			r.log.WithError(err).Error("persist system state")
		}

		err = r.ensureProjectHasLabels(ctx, teamProject, map[string]string{
			"team":                         string(input.Team.Slug),
			"environment":                  environment,
			"tenant":                       r.tenantName,
			reconcilers.ManagedByLabelName: reconcilers.ManagedByLabelValue,
		})
		if err != nil {
			return fmt.Errorf("set project labels: %w", err)
		}

		err = r.setTeamProjectBillingInfo(ctx, teamProject, input)
		if err != nil {
			return fmt.Errorf("set project billing info for project %q for team %q in environment %q: %w", teamProject.ProjectId, input.Team.Slug, environment, err)
		}

		cnrmServiceAccount, err := r.getOrCreateProjectCnrmServiceAccount(ctx, input, teamProject.ProjectId)
		if err != nil {
			return fmt.Errorf("create CNRM service account for project %q for team %q in environment %q: %w", teamProject.ProjectId, input.Team.Slug, environment, err)
		}

		err = r.setProjectPermissions(ctx, teamProject, input, *googleWorkspaceState.GroupEmail, cluster.ProjectID, cnrmServiceAccount, environment)
		if err != nil {
			return fmt.Errorf("set group permissions to project %q for team %q in environment %q: %w", teamProject.ProjectId, input.Team.Slug, environment, err)
		}

		err = r.ensureProjectHasAccessToGoogleApis(ctx, teamProject, input)
		if err != nil {
			return fmt.Errorf("enable Google APIs access in project %q for team %q in environment %q: %w", teamProject.ProjectId, input.Team.Slug, environment, err)
		}
	}

	return nil
}

func (r *googleGcpReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	log := r.log.WithTeamSlug(string(teamSlug))
	state := &reconcilers.GoogleGcpProjectState{
		Projects: make(map[string]reconcilers.GoogleGcpEnvironmentProject),
	}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), teamSlug, state)
	if err != nil {
		return fmt.Errorf("load reconciler state for team %q in reconciler %q: %w", teamSlug, r.Name(), err)
	}
	if state.Projects == nil {
		state.Projects = make(map[string]reconcilers.GoogleGcpEnvironmentProject)
	}

	if len(state.Projects) == 0 {
		log.Info("no GCP projects in reconciler state, nothing to delete")
		return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
	}

	var errors []error

	for environment, teamProject := range state.Projects {
		_, exists := r.clusters[environment]
		if !exists {
			log.Error("environment %q is no longer active, removing from state")
			delete(state.Projects, environment)
			continue
		}

		_, err = r.gcpServices.CloudResourceManagerProjectsService.Delete("projects/" + teamProject.ProjectID).Context(ctx).Do()
		if err != nil {
			errors = append(errors, err)
			continue
		}

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(teamSlug),
		}
		fields := auditlogger.Fields{
			Action:        sqlc.AuditActionGoogleGcpDeleteProject,
			CorrelationID: correlationID,
		}
		r.auditLogger.Logf(ctx, r.database, targets, fields, "Delete GCP project: %q", teamProject.ProjectID)
		delete(state.Projects, environment)
	}

	if len(errors) == 0 {
		return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
	}

	metrics.IncDeleteErrorCounter(len(errors))
	for _, err := range errors {
		log.WithError(err).Error("error during team deletion")
	}

	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), teamSlug, state)
	if err != nil {
		log.WithError(err).Error("persist reconciler state during delete")
	}

	return fmt.Errorf("%d error(s) occurred during GCP project deletion", len(errors))
}

func (r *googleGcpReconciler) ensureProjectHasAccessToGoogleApis(ctx context.Context, project *cloudresourcemanager.Project, input reconcilers.Input) error {
	desiredServiceIDs := map[string]struct{}{
		"compute.googleapis.com":              {},
		"cloudbilling.googleapis.com":         {},
		"storage-component.googleapis.com":    {},
		"storage-api.googleapis.com":          {},
		"sqladmin.googleapis.com":             {},
		"sql-component.googleapis.com":        {},
		"cloudresourcemanager.googleapis.com": {},
		"secretmanager.googleapis.com":        {},
		"pubsub.googleapis.com":               {},
		"logging.googleapis.com":              {},
		"bigquery.googleapis.com":             {},
		"cloudtrace.googleapis.com":           {},
	}

	response, err := r.gcpServices.ServiceUsageService.List(project.Name).Filter("state:ENABLED").Do()
	if err != nil {
		metrics.IncExternalCallsByError(metricsSystemName, err)
		return err
	}
	metrics.IncExternalCalls(metricsSystemName, response.HTTPStatusCode)

	if response.HTTPStatusCode != http.StatusOK {
		return fmt.Errorf("non OK http status: %v", response.HTTPStatusCode)
	}

	// Take already enabled services out of the list of services we want to enable
	for _, enabledService := range response.Services {
		delete(desiredServiceIDs, enabledService.Config.Name)
	}

	if len(desiredServiceIDs) == 0 {
		return nil
	}

	servicesToEnable := make([]string, 0, len(desiredServiceIDs))
	for key := range desiredServiceIDs {
		servicesToEnable = append(servicesToEnable, key)
	}

	req := &serviceusage.BatchEnableServicesRequest{
		ServiceIds: servicesToEnable,
	}

	operation, err := r.gcpServices.ServiceUsageService.BatchEnable(project.Name, req).Do()
	if err != nil {
		metrics.IncExternalCallsByError(metricsSystemName, err)
		return err
	}
	metrics.IncExternalCalls(metricsSystemName, operation.HTTPStatusCode)

	for !operation.Done {
		time.Sleep(1 * time.Second)
		operation, err = r.gcpServices.ServiceUsageOperationsService.Get(operation.Name).Do()
		if err != nil {
			metrics.IncExternalCallsByError(metricsSystemName, err)
			return fmt.Errorf("poll operation: %w", err)
		}
		metrics.IncExternalCalls(metricsSystemName, operation.HTTPStatusCode)
	}

	if operation.Error != nil {
		return fmt.Errorf("complete operation: %s", operation.Error.Message)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(input.Team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGoogleGcpProjectEnableGoogleApis,
		CorrelationID: input.CorrelationID,
	}
	for _, enabledApi := range servicesToEnable {
		r.auditLogger.Logf(ctx, r.database, targets, fields, "Enable Google API %q for %q", enabledApi, project.ProjectId)
	}

	return nil
}

func (r *googleGcpReconciler) getOrCreateProject(ctx context.Context, projectID string, state *reconcilers.GoogleGcpProjectState, environment string, parentFolderID int64, input reconcilers.Input) (*cloudresourcemanager.Project, error) {
	if projectFromState, exists := state.Projects[environment]; exists {
		response, err := r.gcpServices.CloudResourceManagerProjectsService.Search().Query("id:" + projectFromState.ProjectID).Do()
		if err != nil {
			metrics.IncExternalCallsByError(metricsSystemName, err)
			return nil, err
		}
		metrics.IncExternalCalls(metricsSystemName, response.HTTPStatusCode)

		if len(response.Projects) == 1 {
			return response.Projects[0], nil
		}

		if len(response.Projects) > 1 {
			return nil, fmt.Errorf("multiple projects with id: %q found, unable to continue", projectFromState.ProjectID)
		}
	}

	project := &cloudresourcemanager.Project{
		DisplayName: GetProjectDisplayName(input.Team.Slug, environment),
		Parent:      "folders/" + strconv.FormatInt(parentFolderID, 10),
		ProjectId:   projectID,
	}
	operation, err := r.gcpServices.CloudResourceManagerProjectsService.Create(project).Do()
	if err != nil {
		googleError, ok := err.(*googleapi.Error)
		if !ok {
			metrics.IncExternalCallsByError(metricsSystemName, err)
			return nil, fmt.Errorf("create GCP project: %w", err)
		}

		if googleError.Code != 409 {
			metrics.IncExternalCallsByError(metricsSystemName, err)
			return nil, fmt.Errorf("create GCP project: %w", err)
		}

		// the project already exists, adopt
		response, err := r.gcpServices.CloudResourceManagerProjectsService.Search().Query("id:" + projectID).Do()
		if err != nil {
			metrics.IncExternalCallsByError(metricsSystemName, err)
			return nil, fmt.Errorf("find existing GCP project: %w", err)
		}

		metrics.IncExternalCalls(metricsSystemName, response.HTTPStatusCode)

		if len(response.Projects) != 1 {
			return nil, fmt.Errorf("invalid number of projects in response: %+v", response.Projects)
		}

		return response.Projects[0], nil

	}
	metrics.IncExternalCalls(metricsSystemName, operation.HTTPStatusCode)

	response, err := r.getOperationResponse(operation)
	if err != nil {
		return nil, fmt.Errorf("get result from GCP project creation: %w", err)
	}

	createdProject := &cloudresourcemanager.Project{}
	err = json.Unmarshal(response, createdProject)
	if err != nil {
		return nil, fmt.Errorf("convert operation response to the Created GCP project: %w", err)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(input.Team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGoogleGcpProjectCreateProject,
		CorrelationID: input.CorrelationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Created GCP project %q for team %q in environment %q", createdProject.ProjectId, input.Team.Slug, environment)

	return createdProject, nil
}

func (r *googleGcpReconciler) getLegacyMember(teamSlug slug.Slug, env string) string {
	if r.tenantName != "nav" {
		return ""
	}
	legacyClusters := map[string]string{
		"dev-gcp":  "nais-dev-2e7b",
		"prod-gcp": "nais-prod-020f",
		"ci-gcp":   "nais-ci-e17f",
	}
	for _, m := range r.legacyMapping {
		if m.Platinum == env && legacyClusters[m.Legacy] != "" {
			return fmt.Sprintf("serviceAccount:%s.svc.id.goog[cnrm-system/cnrm-controller-manager-%s]", legacyClusters[m.Legacy], teamSlug)
		}
	}
	return ""
}

// setProjectPermissions Make sure that the project has the necessary permissions, and don't remove permissions we don't
// control
func (r *googleGcpReconciler) setProjectPermissions(ctx context.Context, project *cloudresourcemanager.Project, input reconcilers.Input, groupEmail, clusterProjectID string, cnrmServiceAccount *iam.ServiceAccount, environment string) error {
	// Set workload identity role to the CNRM service account
	member := fmt.Sprintf("serviceAccount:%s.svc.id.goog[cnrm-system/cnrm-controller-manager-%s]", clusterProjectID, input.Team.Slug)
	members := []string{member}

	member = r.getLegacyMember(input.Team.Slug, environment)
	if member != "" {
		members = append(members, member)
	}

	operation, err := r.gcpServices.IamProjectsServiceAccountsService.SetIamPolicy(cnrmServiceAccount.Name, &iam.SetIamPolicyRequest{
		Policy: &iam.Policy{
			Bindings: []*iam.Binding{
				{
					Members: members,
					Role:    "roles/iam.workloadIdentityUser",
				},
			},
		},
	}).Do()
	if err != nil {
		metrics.IncExternalCallsByError(metricsSystemName, err)
		return fmt.Errorf("assign roles for CNRM service account: %w", err)
	}
	metrics.IncExternalCalls(metricsSystemName, operation.HTTPStatusCode)

	policy, err := r.gcpServices.CloudResourceManagerProjectsService.GetIamPolicy(project.Name, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		metrics.IncExternalCallsByError(metricsSystemName, err)
		return fmt.Errorf("retrieve existing GCP project IAM policy: %w", err)
	}
	metrics.IncExternalCalls(metricsSystemName, policy.HTTPStatusCode)

	newBindings, updated := calculateRoleBindings(policy.Bindings, map[string]string{
		"roles/owner":  "group:" + groupEmail,
		r.cnrmRoleName: "serviceAccount:" + cnrmServiceAccount.Email,
	})

	if !updated {
		return nil
	}

	policy.Bindings = newBindings
	policy, err = r.gcpServices.CloudResourceManagerProjectsService.SetIamPolicy(project.Name, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		metrics.IncExternalCallsByError(metricsSystemName, err)
		return fmt.Errorf("assign GCP project IAM policy: %w", err)
	}
	metrics.IncExternalCalls(metricsSystemName, policy.HTTPStatusCode)

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(input.Team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGoogleGcpProjectAssignPermissions,
		CorrelationID: input.CorrelationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Assigned GCP project IAM permissions for %q", project.ProjectId)

	return nil
}

// getOrCreateProjectCnrmServiceAccount Get the CNRM service account for the project in this env. If the service account
// does not exist, attempt to create it, and then return it.
func (r *googleGcpReconciler) getOrCreateProjectCnrmServiceAccount(ctx context.Context, input reconcilers.Input, teamProjectID string) (*iam.ServiceAccount, error) {
	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", reconcilers.CnrmServiceAccountAccountID, teamProjectID)
	name := fmt.Sprintf("projects/-/serviceAccounts/%s", email)
	serviceAccount, err := r.gcpServices.IamProjectsServiceAccountsService.Get(name).Do()
	if err == nil {
		metrics.IncExternalCalls(metricsSystemName, serviceAccount.HTTPStatusCode)
		return serviceAccount, nil
	}
	metrics.IncExternalCalls(metricsSystemName, 0)

	createServiceAccountRequest := &iam.CreateServiceAccountRequest{
		AccountId: reconcilers.CnrmServiceAccountAccountID,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: "CNRM service account",
			Description: "Managed by Console",
		},
	}
	serviceAccount, err = r.gcpServices.IamProjectsServiceAccountsService.Create("projects/"+teamProjectID, createServiceAccountRequest).Do()
	if err != nil {
		metrics.IncExternalCallsByError(metricsSystemName, err)
		return nil, err
	}
	metrics.IncExternalCalls(metricsSystemName, serviceAccount.HTTPStatusCode)

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(input.Team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGoogleGcpProjectCreateCnrmServiceAccount,
		CorrelationID: input.CorrelationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Created CNRM service account for team %q in project %q", input.Team.Slug, teamProjectID)

	return serviceAccount, nil
}

func (r *googleGcpReconciler) setTeamProjectBillingInfo(ctx context.Context, project *cloudresourcemanager.Project, input reconcilers.Input) error {
	info, err := r.gcpServices.CloudBillingProjectsService.GetBillingInfo(project.Name).Do()
	if err != nil {
		metrics.IncExternalCallsByError(metricsSystemName, err)
		return err
	}
	metrics.IncExternalCalls(metricsSystemName, info.HTTPStatusCode)

	if info.BillingAccountName == r.billingAccount {
		return nil
	}

	updatedBillingInfo, err := r.gcpServices.CloudBillingProjectsService.UpdateBillingInfo(project.Name, &cloudbilling.ProjectBillingInfo{
		BillingAccountName: r.billingAccount,
	}).Do()
	if err != nil {
		metrics.IncExternalCallsByError(metricsSystemName, err)
		return err
	}
	metrics.IncExternalCalls(metricsSystemName, updatedBillingInfo.HTTPStatusCode)

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(input.Team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGoogleGcpProjectSetBillingInfo,
		CorrelationID: input.CorrelationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Set billing info for %q", project.ProjectId)

	return nil
}

func (r *googleGcpReconciler) getOperationResponse(operation *cloudresourcemanager.Operation) (googleapi.RawMessage, error) {
	var err error
	for !operation.Done {
		time.Sleep(1 * time.Second) // Make sure not to hammer the Operation API
		operation, err = r.gcpServices.CloudResourceManagerOperationsService.Get(operation.Name).Do()
		if err != nil {
			metrics.IncExternalCallsByError(metricsSystemName, err)
			return nil, fmt.Errorf("poll operation: %w", err)
		}
		metrics.IncExternalCalls(metricsSystemName, operation.HTTPStatusCode)
	}

	if operation.Error != nil {
		return nil, fmt.Errorf("complete operation: %s", operation.Error.Message)
	}

	return operation.Response, nil
}

func (r *googleGcpReconciler) ensureProjectHasLabels(_ context.Context, project *cloudresourcemanager.Project, labels map[string]string) error {
	operation, err := r.gcpServices.CloudResourceManagerProjectsService.Patch(project.Name, &cloudresourcemanager.Project{
		Labels: labels,
	}).Do()
	if err != nil {
		metrics.IncExternalCallsByError(metricsSystemName, err)
		return err
	}
	metrics.IncExternalCalls(metricsSystemName, operation.HTTPStatusCode)

	_, err = r.getOperationResponse(operation)
	return err
}

// createGcpServices Creates the GCP services used by the reconciler
func createGcpServices(ctx context.Context, cfg *config.Config) (*GcpServices, error) {
	builder, err := google_token_source.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	ts, err := builder.GCP(ctx)
	if err != nil {
		return nil, fmt.Errorf("get delegated token source: %w", err)
	}

	cloudResourceManagerService, err := cloudresourcemanager.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("retrieve cloud resource manager service: %w", err)
	}

	iamService, err := iam.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("retrieve IAM service service: %w", err)
	}

	cloudBillingService, err := cloudbilling.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("retrieve cloud billing service: %w", err)
	}

	serviceUsageService, err := serviceusage.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("retrieve service usage service: %w", err)
	}

	return &GcpServices{
		CloudBillingProjectsService:           cloudBillingService.Projects,
		CloudResourceManagerProjectsService:   cloudResourceManagerService.Projects,
		CloudResourceManagerOperationsService: cloudResourceManagerService.Operations,
		IamProjectsServiceAccountsService:     iamService.Projects.ServiceAccounts,
		ServiceUsageService:                   serviceUsageService.Services,
		ServiceUsageOperationsService:         serviceUsageService.Operations,
	}, nil
}

// GenerateProjectID Generate a unique project ID for the team in a given environment in a deterministic fashion
func GenerateProjectID(domain, environment string, slug slug.Slug) string {
	hasher := sha256.New()
	hasher.Write([]byte(slug))
	hasher.Write([]byte{0})
	hasher.Write([]byte(environment))
	hasher.Write([]byte{0})
	hasher.Write([]byte(domain))

	parts := make([]string, 3)
	parts[0] = strings.TrimSuffix(console.Truncate(string(slug), 20), "-")
	parts[1] = strings.TrimSuffix(console.Truncate(environment, 4), "-")
	parts[2] = console.Truncate(hex.EncodeToString(hasher.Sum(nil)), 4)

	return strings.Join(parts, "-")
}

// GetProjectDisplayName Get the display name of a project for a team in a given environment
func GetProjectDisplayName(slug slug.Slug, environment string) string {
	suffix := "-" + environment
	maxSlugLength := GoogleProjectDisplayNameMaxLength - len(suffix)
	prefix := console.Truncate(string(slug), maxSlugLength)
	prefix = strings.TrimSuffix(prefix, "-")
	return prefix + suffix
}

// calculateRoleBindings Given a set of role bindings, make sure the ones in requiredRoleBindings are present
func calculateRoleBindings(existingRoleBindings []*cloudresourcemanager.Binding, requiredRoleBindings map[string]string) ([]*cloudresourcemanager.Binding, bool) {
	updated := false

REQUIRED:
	for role, member := range requiredRoleBindings {
		for idx, binding := range existingRoleBindings {
			if binding.Role != role {
				continue
			}

			if !contains(binding.Members, member) {
				existingRoleBindings[idx].Members = append(existingRoleBindings[idx].Members, member)
				updated = true
			}

			continue REQUIRED
		}

		// the required role is missing altogether from the existing bindings
		existingRoleBindings = append(existingRoleBindings, &cloudresourcemanager.Binding{
			Members: []string{member},
			Role:    role,
		})
		updated = true
	}

	return existingRoleBindings, updated
}

// contains Check if a specific value is in a slice of strings
func contains(strings []string, contains string) bool {
	for _, value := range strings {
		if value == contains {
			return true
		}
	}
	return false
}
