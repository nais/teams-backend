package google_gcp_reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/cloudbilling/v1"

	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"

	"github.com/nais/console/pkg/slug"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

type Cluster struct {
	TeamFolderID int64  `json:"team_folder_id"`
	ProjectID    string `json:"cluster_project_id"`
}

type ClusterInfo map[string]Cluster

type googleGcpReconciler struct {
	database                    db.Database
	config                      *jwt.Config
	domain                      string
	auditLogger                 auditlogger.AuditLogger
	clusters                    ClusterInfo
	cnrmRoleName                string
	cloudResourceManagerService *cloudresourcemanager.Service
	iamService                  *iam.Service
	cloudbillingService         *cloudbilling.APIService
	billingAccount              string
}

const (
	Name = sqlc.SystemNameGoogleGcpProject
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, domain string, config *jwt.Config, clusters ClusterInfo, cnrmRoleName string, crmService *cloudresourcemanager.Service, iamService *iam.Service, cloudBilling *cloudbilling.APIService, billingAccount string) *googleGcpReconciler {
	return &googleGcpReconciler{
		database:                    database,
		auditLogger:                 auditLogger,
		domain:                      domain,
		config:                      config,
		clusters:                    clusters,
		cnrmRoleName:                cnrmRoleName,
		cloudResourceManagerService: crmService,
		iamService:                  iamService,
		cloudbillingService:         cloudBilling,
		billingAccount:              billingAccount,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	if !cfg.GCP.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	b, err := os.ReadFile(cfg.Google.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("read google credentials file: %w", err)
	}

	cf, err := google.JWTConfigFromJSON(
		b,
		cloudresourcemanager.CloudPlatformScope,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize google credentials: %w", err)
	}

	client := cf.Client(ctx)

	crmService, err := cloudresourcemanager.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("retrieve cloud resource manager client: %w", err)
	}

	iamService, err := iam.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("retrieve IAM service client: %w", err)
	}

	cloudBillingService, err := cloudbilling.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("retrieve cloud billing service client: %w", err)
	}

	clusters := ClusterInfo{}
	err = json.NewDecoder(strings.NewReader(cfg.GCP.Clusters)).Decode(&clusters)
	if err != nil {
		return nil, fmt.Errorf("parse GCP cluster info: %w", err)
	}

	return New(database, auditLogger, cfg.TenantDomain, cf, clusters, cfg.GCP.CnrmRole, crmService, iamService, cloudBillingService, cfg.GCP.BillingAccount), nil
}

func (r *googleGcpReconciler) Name() sqlc.SystemName {
	return Name
}

func (r *googleGcpReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.GoogleGcpProjectState{
		Projects: make(map[string]reconcilers.GoogleGcpEnvironmentProject),
	}
	err := r.database.LoadSystemState(ctx, r.Name(), input.Team.ID, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	googleWorkspaceState := &reconcilers.GoogleWorkspaceState{}
	err = r.database.LoadSystemState(ctx, google_workspace_admin_reconciler.Name, input.Team.ID, googleWorkspaceState)
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

		err = r.database.SetSystemState(ctx, r.Name(), input.Team.ID, state)
		if err != nil {
			log.Errorf("system state not persisted: %s", err)
		}

		err = r.setProjectPermissions(ctx, project, input, *googleWorkspaceState.GroupEmail, environment, cluster)
		if err != nil {
			return fmt.Errorf("unable to set group permissions to project %q for team %q in environment %q: %w", project.Name, input.Team.Slug, environment, err)
		}

		err = r.setProjectBillingInfo(project.Name)
		if err != nil {
			return fmt.Errorf("unable to set project billing info for project %q for team %q in environment %q: %w", project.Name, input.Team.Slug, environment, err)
		}
	}

	return nil
}

func (r *googleGcpReconciler) getOrCreateProject(ctx context.Context, state *reconcilers.GoogleGcpProjectState, environment string, parentFolderID int64, input reconcilers.Input) (*cloudresourcemanager.Project, error) {
	if projectFromState, exists := state.Projects[environment]; exists {
		project, err := r.cloudResourceManagerService.Projects.Get(projectFromState.ProjectName).Do()
		if err == nil {
			return project, nil
		}
	}

	projectId := GenerateProjectID(r.domain, environment, input.Team.Slug)
	project := &cloudresourcemanager.Project{
		DisplayName: input.Team.Name,
		Parent:      "folders/" + strconv.FormatInt(parentFolderID, 10),
		ProjectId:   projectId,
	}
	operation, err := r.cloudResourceManagerService.Projects.Create(project).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create GCP project: %w", err)
	}

	for !operation.Done {
		time.Sleep(1 * time.Second) // Make sure not to hammer the Operation API
		operation, err = r.cloudResourceManagerService.Operations.Get(operation.Name).Do()
		if err != nil {
			return nil, fmt.Errorf("unable to poll GCP project creation: %w", err)
		}
	}

	if operation.Error != nil {
		return nil, fmt.Errorf("unable to create GCP project: %s", operation.Error.Message)
	}

	createdProject := &cloudresourcemanager.Project{}
	err = json.Unmarshal(operation.Response, createdProject)
	if err != nil {
		return nil, fmt.Errorf("unable to convert operation response to the created GCP project: %w", err)
	}

	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGoogleGcpProjectCreateProject,
		CorrelationID:  input.CorrelationID,
		TargetTeamSlug: &input.Team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "created GCP project %q for team %q in environment %q", createdProject.Name, input.Team.Slug, environment)

	return createdProject, nil
}

// setProjectPermissions Give owner permissions to the team group. The group is created by the Google Workspace Admin
// reconciler. projectName is in the "projects/{ProjectIdOrNumber}" format, and not the project ID
func (r *googleGcpReconciler) setProjectPermissions(ctx context.Context, project *cloudresourcemanager.Project, input reconcilers.Input, groupEmail, environment string, cluster Cluster) error {
	cnrmServiceAccount, err := r.getOrCreateProjectCnrmServiceAccount(ctx, input.Team.Slug, environment, cluster)
	if err != nil {
		return fmt.Errorf("unable to create CNRM service account for project %q for team %q in environment %q: %w", project.Name, input.Team.Slug, environment, err)
	}

	// Set workload identity role to the CNRM service account
	member := fmt.Sprintf("serviceAccount:%s.svc.id.goog[cnrm-system/cnrm-controller-manager-%s]", cluster.ProjectID, input.Team.Slug)
	_, err = r.iamService.Projects.ServiceAccounts.SetIamPolicy(cnrmServiceAccount.Name, &iam.SetIamPolicyRequest{
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

	// Set worksppace group roles
	_, err = r.cloudResourceManagerService.Projects.SetIamPolicy(project.Name, &cloudresourcemanager.SetIamPolicyRequest{
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
	r.auditLogger.Logf(ctx, fields, "assigned GCP project IAM permissions for %q", project.Name)

	return nil
}

// getOrCreateProjectCnrmServiceAccount Get the CNRM service account for the project in this env. If the service account
// does not exist, attempt to create it, and then return it.
func (r *googleGcpReconciler) getOrCreateProjectCnrmServiceAccount(_ context.Context, slug slug.Slug, environment string, cluster Cluster) (*iam.ServiceAccount, error) {
	name, accountID := cnrmServiceAccountNameAndAccountID(slug, cluster.ProjectID)
	serviceAccount, err := r.iamService.Projects.ServiceAccounts.Get(name).Do()
	if err == nil {
		return serviceAccount, nil
	}

	createServiceAccontRequest := &iam.CreateServiceAccountRequest{
		AccountId: accountID,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: fmt.Sprintf("%s CNRM service account (%s)", slug, environment),
			Description: fmt.Sprintf("CNRM service account for team %q in environment %q", slug, environment),
		},
	}
	serviceAccount, err = r.iamService.Projects.ServiceAccounts.Create("projects/"+cluster.ProjectID, createServiceAccontRequest).Do()
	if err != nil {
		return nil, err
	}
	return serviceAccount, nil
}

func (r *googleGcpReconciler) setProjectBillingInfo(projectName string) error {
	_, err := r.cloudbillingService.Projects.UpdateBillingInfo(projectName, &cloudbilling.ProjectBillingInfo{
		BillingAccountName: r.billingAccount,
	}).Do()
	return err
}

// cnrmServiceAccountNameAndAccountID Generate a name and an account ID for a CNRM service account
func cnrmServiceAccountNameAndAccountID(slug slug.Slug, projectID string) (name, accountID string) {
	accountID = fmt.Sprintf("cnrm-%s", slug)
	cnrmEmailAddress := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountID, projectID)
	name = "projects/" + projectID + "/serviceAccounts/" + cnrmEmailAddress
	return
}
