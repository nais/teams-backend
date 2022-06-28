package google_gcp_reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type googleGcpReconciler struct {
	db               *gorm.DB
	config           *jwt.Config
	domain           string
	auditLogger      auditlogger.AuditLogger
	projectParentIDs map[string]int64
	system           dbmodels.System
}

const (
	Name                = "google:gcp:project"
	OpCreateProject     = "google:gcp:project:create-project"
	OpAssignPermissions = "google:gcp:project:assign-permissions"
)

func New(db *gorm.DB, system dbmodels.System, auditLogger auditlogger.AuditLogger, domain string, config *jwt.Config, projectParentIDs map[string]int64) *googleGcpReconciler {
	return &googleGcpReconciler{
		db:               db,
		auditLogger:      auditLogger,
		domain:           domain,
		config:           config,
		projectParentIDs: projectParentIDs,
		system:           system,
	}
}

func NewFromConfig(db *gorm.DB, cfg *config.Config, system dbmodels.System, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
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

	return New(db, system, auditLogger, cfg.TenantDomain, cf, cfg.GCP.ProjectParentIDs), nil
}

func (r *googleGcpReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.GoogleGcpProjectState{
		Projects: make(map[string]reconcilers.GoogleGcpEnvironmentProject),
	}
	err := dbmodels.LoadSystemState(r.db, *r.system.ID, *input.Team.ID, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team '%s' in system '%s': %w", input.Team.Slug, r.system.Name, err)
	}

	client := r.config.Client(ctx)
	svc, err := cloudresourcemanager.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("retrieve cloud resource manager client: %w", err)
	}

	for environment, parentFolderID := range r.projectParentIDs {
		project, err := r.getOrCreateProject(svc, state, environment, parentFolderID, input.Corr, input.Team)
		if err != nil {
			return fmt.Errorf("unable to get or create a GCP project for team '%s' in environment '%s': %w", input.Team.Slug, environment, err)
		}
		state.Projects[environment] = reconcilers.GoogleGcpEnvironmentProject{
			ProjectID:   project.ProjectId,
			ProjectName: project.Name,
		}
		err = dbmodels.SetSystemState(r.db, *r.system.ID, *input.Team.ID, state)
		if err != nil {
			log.Errorf("system state not persisted: %s", err)
		}

		err = r.setProjectPermissions(svc, project.Name, input.Corr, input.Team)
		if err != nil {
			return fmt.Errorf("unable to set group permissions to project '%s' for team '%s' in environment '%s': %w", project.Name, input.Team.Slug, environment, err)
		}
	}

	return nil
}

func (r *googleGcpReconciler) System() dbmodels.System {
	return r.system
}

func (r *googleGcpReconciler) getOrCreateProject(svc *cloudresourcemanager.Service, state *reconcilers.GoogleGcpProjectState, environment string, parentFolderID int64, corr dbmodels.Correlation, team dbmodels.Team) (*cloudresourcemanager.Project, error) {
	if projectFromState, exists := state.Projects[environment]; exists {
		project, err := svc.Projects.Get(projectFromState.ProjectName).Do()
		if err == nil {
			return project, nil
		}
	}

	projectId := GenerateProjectID(r.domain, environment, string(team.Slug))
	project := &cloudresourcemanager.Project{
		DisplayName: team.Name,
		Parent:      "folders/" + strconv.FormatInt(parentFolderID, 10),
		ProjectId:   projectId,
	}
	operation, err := svc.Projects.Create(project).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create GCP project: %w", err)
	}

	for !operation.Done {
		time.Sleep(1 * time.Second) // Make sure not to hammer the Operation API
		operation, err = svc.Operations.Get(operation.Name).Do()
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

	r.auditLogger.Logf(OpCreateProject, corr, r.system, nil, &team, nil, "created GCP project '%s' for team '%s' in environment '%s'", createdProject.Name, team.Slug, environment)

	return createdProject, nil
}

// createPermissions Give owner permissions to the team group. The group is created by the Google Workspace Admin
// reconciler. projectName is in the "projects/{ProjectIdOrNumber}" format, and not the project ID
func (r *googleGcpReconciler) setProjectPermissions(svc *cloudresourcemanager.Service, projectName string, corr dbmodels.Correlation, team dbmodels.Team) error {
	// FIXME: Check state to make sure we are generating the correct group name
	member := fmt.Sprintf("group:%s%s@%s", reconcilers.TeamNamePrefix, team.Slug, r.domain)
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
	_, err := svc.Projects.SetIamPolicy(projectName, req).Do()
	if err != nil {
		return fmt.Errorf("assign GCP project IAM permissions: %w", err)
	}

	r.auditLogger.Logf(OpAssignPermissions, corr, r.system, nil, &team, nil, "assigned GCP project IAM permissions for '%s'", projectName)

	return nil
}
