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

func New(database db.Database, auditLogger auditlogger.AuditLogger, clusters gcp.Clusters, gcpServices *GcpServices, tenantName, domain, cnrmRoleName, billingAccount string, legacyClusters map[string]string, legacyMapping []envmap.EnvironmentMapping, log logger.Logger) *googleGcpReconciler {
	return &googleGcpReconciler{
		database:       database,
		auditLogger:    auditLogger,
		clusters:       clusters,
		gcpServices:    gcpServices,
		domain:         domain,
		cnrmRoleName:   cnrmRoleName,
		billingAccount: billingAccount,
		tenantName:     tenantName,
		legacyClusters: legacyClusters,
		legacyMapping:  legacyMapping,
		log:            log,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	log = log.WithSystem(string(Name))

	gcpServices, err := createGcpServices(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return New(database, auditLogger, cfg.GCP.Clusters, gcpServices, cfg.TenantName, cfg.TenantDomain, cfg.GCP.CnrmRole, cfg.GCP.BillingAccount, cfg.LegacyClusters, cfg.LegacyNaisNamespaces, log), nil
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
		project, err := r.getOrCreateProject(ctx, projectID, state, environment, cluster.TeamsFolderID, input)
		if err != nil {
			return fmt.Errorf("get or create a GCP project %q for team %q in environment %q: %w", projectID, input.Team.Slug, environment, err)
		}
		teamProjects[environment] = project
		state.Projects[environment] = reconcilers.GoogleGcpEnvironmentProject{
			ProjectID: project.ProjectId,
		}

		err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, state)
		if err != nil {
			r.log.WithError(err).Error("persist system state")
		}

		err = r.ensureProjectHasLabels(ctx, project, map[string]string{
			"team":        string(input.Team.Slug),
			"environment": environment,
			"tenant":      r.tenantName,
		})
		if err != nil {
			return fmt.Errorf("set project labels: %w", err)
		}

		err = r.setTeamProjectBillingInfo(ctx, project, input)
		if err != nil {
			return fmt.Errorf("set project billing info for project %q for team %q in environment %q: %w", project.ProjectId, input.Team.Slug, environment, err)
		}

		cnrmServiceAccount, err := r.getOrCreateProjectCnrmServiceAccount(ctx, input, environment, cluster.ProjectID)
		if err != nil {
			return fmt.Errorf("create CNRM service account for project %q for team %q in environment %q: %w", project.ProjectId, input.Team.Slug, environment, err)
		}

		err = r.setProjectPermissions(ctx, project, input, *googleWorkspaceState.GroupEmail, cluster.ProjectID, cnrmServiceAccount)
		if err != nil {
			return fmt.Errorf("set group permissions to project %q for team %q in environment %q: %w", project.ProjectId, input.Team.Slug, environment, err)
		}

		err = r.ensureProjectHasAccessToGoogleApis(ctx, project, input)
		if err != nil {
			return fmt.Errorf("enable Google APIs access in project %q for team %q in environment %q: %w", project.ProjectId, input.Team.Slug, environment, err)
		}
	}

	err = r.createLegacyClusterCNRMServiceAccount(ctx, input, teamProjects, *googleWorkspaceState.GroupEmail)
	if err != nil {
		return fmt.Errorf("hack for NAVs legacy GCP projects: %w", err)
	}

	return nil
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

func (r *googleGcpReconciler) createLegacyClusterCNRMServiceAccount(ctx context.Context, input reconcilers.Input, teamProjects map[string]*cloudresourcemanager.Project, groupEmail string) error {
OUTER:
	for legacyEnvironment, legacyClusterProject := range r.legacyClusters {
		var teamProject *cloudresourcemanager.Project
		for _, m := range r.legacyMapping {
			if m.Legacy == legacyEnvironment {
				teamProject = teamProjects[m.Platinum]
				break
			}
		}
		if teamProject == nil {
			r.log.Warnf("found no team project for cluster: %q. legacyClusters: %+v, legacyMapping: %+v, teamProjects: %+v", legacyEnvironment, r.legacyClusters, r.legacyMapping, teamProjects)
			continue OUTER
		}

		cnrmServiceAccount, err := r.getOrCreateProjectCnrmServiceAccount(ctx, input, legacyEnvironment, legacyClusterProject)
		if err != nil {
			return fmt.Errorf("create legacy CNRM service account for team %q in environment %q: %w", input.Team.Slug, legacyEnvironment, err)
		}

		// Set workload identity role to the CNRM service account
		member := fmt.Sprintf("serviceAccount:%s.svc.id.goog[cnrm-system/cnrm-controller-manager-%s]", legacyClusterProject, input.Team.Slug)
		response, err := r.gcpServices.IamProjectsServiceAccountsService.SetIamPolicy(cnrmServiceAccount.Name, &iam.SetIamPolicyRequest{
			Policy: &iam.Policy{
				Bindings: []*iam.Binding{
					{
						Members: []string{member},
						Role:    "roles/iam.workloadIdentityUser",
					},
				},
			},
		}).Do()
		if err != nil {
			metrics.IncExternalCallsByError(metricsSystemName, err)
			return fmt.Errorf("assign roles for legacy CNRM service account: %w", err)
		}
		metrics.IncExternalCalls(metricsSystemName, response.HTTPStatusCode)

		err = r.setProjectPermissions(ctx, teamProject, input, groupEmail, legacyClusterProject, cnrmServiceAccount)
		if err != nil {
			return fmt.Errorf("set group permissions to project %q for team %q in environment %q: %w", teamProject, input.Team.Slug, legacyEnvironment, err)
		}
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

// setProjectPermissions Make sure that the project has the necessary permissions, and don't remove permissions we don't
// control
func (r *googleGcpReconciler) setProjectPermissions(ctx context.Context, project *cloudresourcemanager.Project, input reconcilers.Input, groupEmail, clusterProjectID string, cnrmServiceAccount *iam.ServiceAccount) error {
	// Set workload identity role to the CNRM service account
	member := fmt.Sprintf("serviceAccount:%s.svc.id.goog[cnrm-system/cnrm-controller-manager-%s]", clusterProjectID, input.Team.Slug)
	operation, err := r.gcpServices.IamProjectsServiceAccountsService.SetIamPolicy(cnrmServiceAccount.Name, &iam.SetIamPolicyRequest{
		Policy: &iam.Policy{
			Bindings: []*iam.Binding{
				{
					Members: []string{member},
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
func (r *googleGcpReconciler) getOrCreateProjectCnrmServiceAccount(ctx context.Context, input reconcilers.Input, environment string, clusterProjectID string) (*iam.ServiceAccount, error) {
	name, accountID := CnrmServiceAccountNameAndAccountID(input.Team.Slug, clusterProjectID)
	serviceAccount, err := r.gcpServices.IamProjectsServiceAccountsService.Get(name).Do()
	if err == nil {
		metrics.IncExternalCalls(metricsSystemName, serviceAccount.HTTPStatusCode)
		return serviceAccount, nil
	}
	metrics.IncExternalCalls(metricsSystemName, 0)

	createServiceAccountRequest := &iam.CreateServiceAccountRequest{
		AccountId: accountID,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: fmt.Sprintf("%s CNRM service account (%s)", input.Team.Slug, environment),
			Description: fmt.Sprintf("CNRM service account for team %q in environment %q", input.Team.Slug, environment),
		},
	}
	serviceAccount, err = r.gcpServices.IamProjectsServiceAccountsService.Create("projects/"+clusterProjectID, createServiceAccountRequest).Do()
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
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Created CNRM service account for team %q in environment %q", input.Team.Slug, environment)

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

// CnrmServiceAccountNameAndAccountID Generate a name and an account ID for a CNRM service account
func CnrmServiceAccountNameAndAccountID(slug slug.Slug, projectID string) (name, accountID string) {
	accountID = console.SlugHashPrefixTruncate(slug, "cnrm", gcp.GoogleServiceAccountMaxLength)
	cnrmEmailAddress := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountID, projectID)
	name = "projects/" + projectID + "/serviceAccounts/" + cnrmEmailAddress
	return
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
