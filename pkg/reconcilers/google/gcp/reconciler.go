package google_gcp_reconciler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

type cluster struct {
	TeamFolderID int64  `json:"team_folder_id"`
	ProjectID    string `json:"cluster_project_id"`
}

type clusterInfo map[string]cluster

type gcpServices struct {
	cloudBilling         *cloudbilling.APIService
	cloudResourceManager *cloudresourcemanager.Service
	iam                  *iam.Service
}

type googleGcpReconciler struct {
	database       db.Database
	auditLogger    auditlogger.AuditLogger
	clusters       clusterInfo
	gcpServices    *gcpServices
	tenantName     string
	domain         string
	cnrmRoleName   string
	billingAccount string
}

const (
	Name = sqlc.SystemNameGoogleGcpProject
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, clusters clusterInfo, gcpServices *gcpServices, tenantName, domain, cnrmRoleName, billingAccount string) *googleGcpReconciler {
	return &googleGcpReconciler{
		database:       database,
		auditLogger:    auditLogger,
		clusters:       clusters,
		gcpServices:    gcpServices,
		domain:         domain,
		cnrmRoleName:   cnrmRoleName,
		billingAccount: billingAccount,
		tenantName:     tenantName,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	if !cfg.GCP.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	gcpServices, err := createGcpServices(ctx, cfg.Google.CredentialsFile)
	if err != nil {
		return nil, err
	}

	clusters := clusterInfo{}
	err = json.NewDecoder(strings.NewReader(cfg.GCP.Clusters)).Decode(&clusters)
	if err != nil {
		return nil, fmt.Errorf("parse GCP cluster info: %w", err)
	}

	return New(database, auditLogger, clusters, gcpServices, cfg.TenantName, cfg.TenantDomain, cfg.GCP.CnrmRole, cfg.GCP.BillingAccount), nil
}

func (r *googleGcpReconciler) Name() sqlc.SystemName {
	return Name
}

func (r *googleGcpReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.GoogleGcpProjectState{
		Projects: make(map[string]reconcilers.GoogleGcpEnvironmentProject),
	}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.ID, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	googleWorkspaceState := &reconcilers.GoogleWorkspaceState{}
	err = r.database.SetReconcilerStateForTeam(ctx, google_workspace_admin_reconciler.Name, input.Team.ID, googleWorkspaceState)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, google_workspace_admin_reconciler.Name, err)
	}

	if googleWorkspaceState.GroupEmail == nil {
		return fmt.Errorf("no Google Workspace group exists for team %q yet, is the %q reconciler enabled? ", input.Team.Slug, google_workspace_admin_reconciler.Name)
	}

	for environment, cluster := range r.clusters {
		project, err := r.getOrCreateProject(ctx, state, environment, cluster.TeamFolderID, input)
		if err != nil {
			return fmt.Errorf("unable to get or create a GCP project for team %q in environment %q: %w", input.Team.Slug, environment, err)
		}
		state.Projects[environment] = reconcilers.GoogleGcpEnvironmentProject{
			ProjectID:   project.ProjectId,
			ProjectName: project.Name,
		}

		err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.ID, state)
		if err != nil {
			log.Errorf("system state not persisted: %s", err)
		}

		err = r.ensureProjectHasLabels(ctx, project, map[string]string{
			"team":        string(input.Team.Slug),
			"environment": environment,
			"tenant":      r.tenantName,
		})
		if err != nil {
			return fmt.Errorf("unable to set project labels: %w", err)
		}

		err = r.setProjectPermissions(ctx, project, input, *googleWorkspaceState.GroupEmail, environment, cluster)
		if err != nil {
			return fmt.Errorf("unable to set group permissions to project %q for team %q in environment %q: %w", project.ProjectId, input.Team.Slug, environment, err)
		}

		err = r.setTeamProjectBillingInfo(ctx, project, input)
		if err != nil {
			return fmt.Errorf("unable to set project billing info for project %q for team %q in environment %q: %w", project.ProjectId, input.Team.Slug, environment, err)
		}
	}

	return nil
}

func (r *googleGcpReconciler) getOrCreateProject(ctx context.Context, state *reconcilers.GoogleGcpProjectState, environment string, parentFolderID int64, input reconcilers.Input) (*cloudresourcemanager.Project, error) {
	if projectFromState, exists := state.Projects[environment]; exists {
		project, err := r.gcpServices.cloudResourceManager.Projects.Get(projectFromState.ProjectName).Do()
		if err == nil {
			return project, nil
		}
	}

	projectID := GenerateProjectID(r.domain, environment, input.Team.Slug)
	project := &cloudresourcemanager.Project{
		DisplayName: string(input.Team.Slug),
		Parent:      "folders/" + strconv.FormatInt(parentFolderID, 10),
		ProjectId:   projectID,
	}
	operation, err := r.gcpServices.cloudResourceManager.Projects.Create(project).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create GCP project: %w", err)
	}

	response, err := r.getOperationResponse(operation)
	if err != nil {
		return nil, fmt.Errorf("unable to create GCP project: %w", err)
	}

	createdProject := &cloudresourcemanager.Project{}
	err = json.Unmarshal(response, createdProject)
	if err != nil {
		return nil, fmt.Errorf("unable to convert operation response to the created GCP project: %w", err)
	}

	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGoogleGcpProjectCreateProject,
		CorrelationID:  input.CorrelationID,
		TargetTeamSlug: &input.Team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "created GCP project %q for team %q in environment %q", createdProject.ProjectId, input.Team.Slug, environment)

	return createdProject, nil
}

// setProjectPermissions Give owner permissions to the team group. The group is created by the Google Workspace Admin
// reconciler. projectName is in the "projects/{ProjectIdOrNumber}" format, and not the project ID
func (r *googleGcpReconciler) setProjectPermissions(ctx context.Context, project *cloudresourcemanager.Project, input reconcilers.Input, groupEmail, environment string, cluster cluster) error {
	cnrmServiceAccount, err := r.getOrCreateProjectCnrmServiceAccount(ctx, input, environment, cluster)
	if err != nil {
		return fmt.Errorf("unable to create CNRM service account for project %q for team %q in environment %q: %w", project.ProjectId, input.Team.Slug, environment, err)
	}

	// Set workload identity role to the CNRM service account
	member := fmt.Sprintf("serviceAccount:%s.svc.id.goog[cnrm-system/cnrm-controller-manager-%s]", cluster.ProjectID, input.Team.Slug)
	_, err = r.gcpServices.iam.Projects.ServiceAccounts.SetIamPolicy(cnrmServiceAccount.Name, &iam.SetIamPolicyRequest{
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
		return fmt.Errorf("assign roles for CNRM service account: %w", err)
	}

	_, err = r.gcpServices.cloudResourceManager.Projects.SetIamPolicy(project.Name, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: &cloudresourcemanager.Policy{
			Bindings: []*cloudresourcemanager.Binding{
				{
					Members: []string{"group:" + groupEmail},
					Role:    "roles/owner",
				},
				{
					Members: []string{"serviceAccount:" + cnrmServiceAccount.Email},
					Role:    r.cnrmRoleName,
				},
			},
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("assign GCP project IAM permissions: %w", err)
	}

	// FIXME: No need to log if no changes are made
	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGoogleGcpProjectAssignPermissions,
		CorrelationID:  input.CorrelationID,
		TargetTeamSlug: &input.Team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "assigned GCP project IAM permissions for %q", project.ProjectId)

	return nil
}

// getOrCreateProjectCnrmServiceAccount Get the CNRM service account for the project in this env. If the service account
// does not exist, attempt to create it, and then return it.
func (r *googleGcpReconciler) getOrCreateProjectCnrmServiceAccount(ctx context.Context, input reconcilers.Input, environment string, cluster cluster) (*iam.ServiceAccount, error) {
	name, accountID := cnrmServiceAccountNameAndAccountID(input.Team.Slug, cluster.ProjectID)
	serviceAccount, err := r.gcpServices.iam.Projects.ServiceAccounts.Get(name).Do()
	if err == nil {
		return serviceAccount, nil
	}

	createServiceAccountRequest := &iam.CreateServiceAccountRequest{
		AccountId: accountID,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: fmt.Sprintf("%s CNRM service account (%s)", input.Team.Slug, environment),
			Description: fmt.Sprintf("CNRM service account for team %q in environment %q", input.Team.Slug, environment),
		},
	}
	serviceAccount, err = r.gcpServices.iam.Projects.ServiceAccounts.Create("projects/"+cluster.ProjectID, createServiceAccountRequest).Do()
	if err != nil {
		return nil, err
	}

	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGoogleGcpProjectCreateCnrmServiceAccount,
		CorrelationID:  input.CorrelationID,
		TargetTeamSlug: &input.Team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "create CNRM service account for team %q in environment %q", input.Team.Slug, environment)

	return serviceAccount, nil
}

func (r *googleGcpReconciler) setTeamProjectBillingInfo(ctx context.Context, project *cloudresourcemanager.Project, input reconcilers.Input) error {
	info, err := r.gcpServices.cloudBilling.Projects.GetBillingInfo(project.Name).Do()
	if err != nil {
		return err
	}

	if info.BillingAccountName == r.billingAccount {
		return nil
	}

	_, err = r.gcpServices.cloudBilling.Projects.UpdateBillingInfo(project.Name, &cloudbilling.ProjectBillingInfo{
		BillingAccountName: r.billingAccount,
	}).Do()
	if err != nil {
		return err
	}

	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGoogleGcpProjectSetBillingInfo,
		CorrelationID:  input.CorrelationID,
		TargetTeamSlug: &input.Team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "set billing info for %q", project.ProjectId)

	return nil
}

func (r *googleGcpReconciler) getOperationResponse(operation *cloudresourcemanager.Operation) (googleapi.RawMessage, error) {
	var err error
	for !operation.Done {
		time.Sleep(1 * time.Second) // Make sure not to hammer the Operation API
		operation, err = r.gcpServices.cloudResourceManager.Operations.Get(operation.Name).Do()
		if err != nil {
			return nil, fmt.Errorf("unable to poll operation: %w", err)
		}
	}

	if operation.Error != nil {
		return nil, fmt.Errorf("unable to complete operation: %s", operation.Error.Message)
	}

	return operation.Response, nil
}

func (r *googleGcpReconciler) ensureProjectHasLabels(_ context.Context, project *cloudresourcemanager.Project, labels map[string]string) error {
	operation, err := r.gcpServices.cloudResourceManager.Projects.Patch(project.Name, &cloudresourcemanager.Project{
		Labels: labels,
	}).Do()
	if err != nil {
		return err
	}

	_, err = r.getOperationResponse(operation)
	return err
}

// cnrmServiceAccountNameAndAccountID Generate a name and an account ID for a CNRM service account
func cnrmServiceAccountNameAndAccountID(slug slug.Slug, projectID string) (name, accountID string) {
	accountID = fmt.Sprintf("cnrm-%s", slug)
	cnrmEmailAddress := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountID, projectID)
	name = "projects/" + projectID + "/serviceAccounts/" + cnrmEmailAddress
	return
}

// createGcpServices Creates the GCP services used by the reconciler
func createGcpServices(ctx context.Context, credentialsFilePath string) (*gcpServices, error) {
	credentials, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("read google credentials file: %w", err)
	}

	jwtConfig, err := google.JWTConfigFromJSON(credentials, cloudresourcemanager.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("initialize google credentials: %w", err)
	}

	client := jwtConfig.Client(ctx)

	cloudResourceManagerService, err := cloudresourcemanager.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("retrieve cloud resource manager service: %w", err)
	}

	iamService, err := iam.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("retrieve IAM service service: %w", err)
	}

	cloudBillingService, err := cloudbilling.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("retrieve cloud billing service: %w", err)
	}

	return &gcpServices{
		cloudBilling:         cloudBillingService,
		cloudResourceManager: cloudResourceManagerService,
		iam:                  iamService,
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
	parts[0] = console.Truncate(string(slug), 20)
	parts[1] = console.Truncate(environment, 4)
	parts[2] = console.Truncate(hex.EncodeToString(hasher.Sum(nil)), 4)

	return strings.Join(parts, "-")
}
