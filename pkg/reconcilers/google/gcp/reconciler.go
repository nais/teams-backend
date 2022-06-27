package google_gcp_reconciler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/googleapi"
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

	return New(db, system, auditLogger, cfg.PartnerDomain, cf, cfg.GCP.ProjectParentIDs), nil
}

func (r *googleGcpReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	client := r.config.Client(ctx)
	svc, err := cloudresourcemanager.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("retrieve cloud resource manager client: %w", err)
	}

	for environment, parentID := range r.projectParentIDs {
		proj, err := r.CreateProject(svc, environment, parentID, input.Corr, input.Team)
		if err != nil {
			return err
		}

		/*
			err = s.saveProjectState(input.Team.ID, environment, proj.ProjectId)
			if err != nil {
				return fmt.Errorf("%s: create GCP project: project was created, but ID could not be stored in database: %w", OpCreateProject, err)
			}
		*/

		err = r.CreatePermissions(svc, proj.Name, input.Corr, input.Team)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *googleGcpReconciler) System() dbmodels.System {
	return r.system
}

func (r *googleGcpReconciler) CreateProject(svc *cloudresourcemanager.Service, environment string, parentID int64, corr dbmodels.Correlation, team dbmodels.Team) (*cloudresourcemanager.Project, error) {
	projectID := CreateProjectID(r.domain, environment, team.Slug.String())

	proj := &cloudresourcemanager.Project{
		Parent:    "folders/" + strconv.FormatInt(parentID, 10),
		ProjectId: projectID,
	}

	operation, err := svc.Projects.Create(proj).Do()

	switch typedError := err.(type) {
	case *googleapi.Error:
		// conflict may be due to
		// 1) already created by us in this folder, or
		// 2) someone else owns this project
		if typedError.Code != http.StatusConflict {
			return nil, fmt.Errorf("%s: create GCP project: %w", OpCreateProject, err)
		}

		query, err := svc.Projects.Search().Query("id:" + projectID).Do()
		if err != nil {
			return nil, fmt.Errorf("%s: create GCP project: %w", OpCreateProject, err)
		}

		if len(query.Projects) == 0 {
			return nil, fmt.Errorf("%s: create GCP project: globally unique project ID is already assigned", OpCreateProject)
		}

		for _, proj = range query.Projects {
			if proj.ProjectId == projectID {
				return proj, nil
			}
		}

		return nil, fmt.Errorf("%s: create GCP project: BUG: search results for project ID returned project without correct ID", OpCreateProject)

	case nil:
		for !operation.Done {
			var err error
			operation, err = svc.Operations.Get(operation.Name).Do()
			if err != nil {
				return nil, fmt.Errorf("%s: create GCP project: %w", OpCreateProject, err)
			}
		}

		if operation.Error != nil {
			return nil, fmt.Errorf("%s: create GCP project: %s", OpCreateProject, operation.Error.Message)
		}

	default:
		return nil, err
	}

	r.auditLogger.Logf(OpCreateProject, corr, r.system, nil, &team, nil, "created GCP project '%s'", proj.Name)

	return proj, nil
}

func (r *googleGcpReconciler) CreatePermissions(svc *cloudresourcemanager.Service, projectName string, corr dbmodels.Correlation, team dbmodels.Team) error {
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

	_, err := svc.Projects.SetIamPolicy(projectName, req).Do()

	if err != nil {
		return fmt.Errorf("%s: assign GCP project IAM permissions: %w", OpAssignPermissions, err)
	}

	r.auditLogger.Logf(OpAssignPermissions, corr, r.system, nil, &team, nil, "assigned GCP project IAM permissions for '%s'", projectName)

	return nil
}
