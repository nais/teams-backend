package google_gcp_reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/option"
)

type googleGcpReconciler struct {
	database                    db.Database
	config                      *jwt.Config
	domain                      string
	auditLogger                 auditlogger.AuditLogger
	projectParentIDs            map[string]int64
	cloudResourceManagerService *cloudresourcemanager.Service
}

const (
	Name = sqlc.SystemNameGoogleGcpProject
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, domain string, config *jwt.Config, projectParentIDs map[string]int64, crmService *cloudresourcemanager.Service) *googleGcpReconciler {
	return &googleGcpReconciler{
		database:                    database,
		auditLogger:                 auditLogger,
		domain:                      domain,
		config:                      config,
		projectParentIDs:            projectParentIDs,
		cloudResourceManagerService: crmService,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	if !cfg.GCP.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	b, err := ioutil.ReadFile(cfg.Google.CredentialsFile)
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

	return New(database, auditLogger, cfg.TenantDomain, cf, cfg.GCP.ProjectParentIDs, crmService), nil
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
		return fmt.Errorf("unable to load system state for team '%s' in system '%s': %w", input.Team.Slug, r.Name(), err)
	}

	for environment, parentFolderID := range r.projectParentIDs {
		project, err := r.getOrCreateProject(ctx, state, environment, parentFolderID, input)
		if err != nil {
			return fmt.Errorf("unable to get or create a GCP project for team '%s' in environment '%s': %w", input.Team.Slug, environment, err)
		}
		state.Projects[environment] = reconcilers.GoogleGcpEnvironmentProject{
			ProjectID:   project.ProjectId,
			ProjectName: project.Name,
		}
		err = r.database.SetSystemState(ctx, r.Name(), input.Team.ID, state)
		if err != nil {
			log.Errorf("system state not persisted: %s", err)
		}

		err = r.setProjectPermissions(ctx, project.Name, input)
		if err != nil {
			return fmt.Errorf("unable to set group permissions to project '%s' for team '%s' in environment '%s': %w", project.Name, input.Team.Slug, environment, err)
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

// createPermissions Give owner permissions to the team group. The group is created by the Google Workspace Admin
// reconciler. projectName is in the "projects/{ProjectIdOrNumber}" format, and not the project ID
func (r *googleGcpReconciler) setProjectPermissions(ctx context.Context, projectName string, input reconcilers.Input) error {
	// FIXME: Check state to make sure we are generating the correct group name
	member := fmt.Sprintf("group:%s%s@%s", reconcilers.TeamNamePrefix, input.Team.Slug, r.domain)
	const owner = "roles/owner"

	req := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: &cloudresourcemanager.Policy{
			Bindings: []*cloudresourcemanager.Binding{
				{
					Members: []string{member},
					Role:    owner,
				},
			},
		},
	}

	// replace all existing policies for the project
	_, err := r.cloudResourceManagerService.Projects.SetIamPolicy(projectName, req).Do()
	if err != nil {
		return fmt.Errorf("assign GCP project IAM permissions: %w", err)
	}

	// FIXME: No need to log if no changes are made
	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGoogleGcpProjectAssignPermissions,
		CorrelationID:  input.CorrelationID,
		TargetTeamSlug: &input.Team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "assigned GCP project IAM permissions for %q", projectName)

	return nil
}
