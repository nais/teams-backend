package google_gcp_reconciler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/registry"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type gcpReconciler struct {
	db               *gorm.DB
	config           *jwt.Config
	domain           string
	logger           auditlogger.Logger
	projectParentIDs map[string]string
}

const (
	Name                = "google:gcp:project"
	OpCreateProject     = "google:gcp:project:create-project"
	OpAssignPermissions = "google:gcp:project:assign-permissions"
)

func init() {
	registry.Register(Name, NewFromConfig)
}

func New(db *gorm.DB, logger auditlogger.Logger, domain string, config *jwt.Config, projectParentIDs map[string]string) *gcpReconciler {
	return &gcpReconciler{
		db:               db,
		logger:           logger,
		domain:           domain,
		config:           config,
		projectParentIDs: projectParentIDs,
	}
}

func NewFromConfig(db *gorm.DB, cfg *config.Config, logger auditlogger.Logger) (reconcilers.Reconciler, error) {
	if !cfg.GCP.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	b, err := ioutil.ReadFile(cfg.GCP.CredentialsFile)
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

	return New(db, logger, cfg.GCP.Domain, cf, cfg.GCP.ProjectParentIDs), nil
}

func (s *gcpReconciler) Reconcile(ctx context.Context, in reconcilers.Input) error {
	client := s.config.Client(ctx)

	svc, err := cloudresourcemanager.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("retrieve cloud resource manager client: %s", err)
	}

	for environment, parentID := range s.projectParentIDs {
		proj, err := s.CreateProject(svc, in, environment, parentID)
		if err != nil {
			return err
		}

		err = saveProjectMeta(s.db, in.Team.ID, environment, proj.ProjectId)
		if err != nil {
			return s.logger.Errorf(in, OpCreateProject, "create GCP project: project was created, but ID could not be stored in database: %s", err)
		}

		err = s.CreatePermissions(svc, in, proj.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *gcpReconciler) CreateProject(svc *cloudresourcemanager.Service, in reconcilers.Input, environment, parentID string) (*cloudresourcemanager.Project, error) {
	projectID := CreateProjectID(s.domain, environment, in.Team.Slug.String())

	proj := &cloudresourcemanager.Project{
		Parent:    parentID,
		ProjectId: projectID,
	}

	oper, err := svc.Projects.Create(proj).Do()

	switch typedError := err.(type) {
	case *googleapi.Error:
		// conflict may be due to
		// 1) already created by us in this folder, or
		// 2) someone else owns this project
		if typedError.Code != http.StatusConflict {
			return nil, s.logger.Errorf(in, OpCreateProject, "create GCP project: %s", err)
		}

		query, err := svc.Projects.Search().Query("id:" + projectID).Do()
		if err != nil {
			return nil, s.logger.Errorf(in, OpCreateProject, "create GCP project: %s", err)
		}

		if len(query.Projects) == 0 {
			return nil, s.logger.Errorf(in, OpCreateProject, "create GCP project: globally unique project ID is already assigned", err)
		}

		for _, proj = range query.Projects {
			if proj.ProjectId == projectID {
				return proj, nil
			}
		}

		return nil, s.logger.Errorf(in, OpCreateProject, "create GCP project: BUG: search results for project ID returned project without correct ID")

	case nil:
		for !oper.Done {
			var err error
			oper, err = svc.Operations.Get(oper.Name).Do()
			if err != nil {
				return nil, s.logger.Errorf(in, OpCreateProject, "create GCP project: %s", err)
			}
		}

		if oper.Error != nil {
			return nil, s.logger.Errorf(in, OpCreateProject, "create GCP project: %s", oper.Error.Message)
		}

	default:
		return nil, err
	}

	s.logger.Logf(in, OpCreateProject, "successfully created GCP project '%s'", proj.Name)

	return proj, nil
}

func (s *gcpReconciler) CreatePermissions(svc *cloudresourcemanager.Service, in reconcilers.Input, projectName string) error {
	member := fmt.Sprintf("group:%s%s@%s", reconcilers.TeamNamePrefix, *in.Team.Slug, s.domain)
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
		return s.logger.Errorf(in, OpAssignPermissions, "assign GCP project IAM permissions: %s", err)
	}

	s.logger.Logf(in, OpAssignPermissions, "successfully assigned GCP project IAM permissions for '%s'", projectName)

	return nil
}

func saveProjectMeta(db *gorm.DB, teamID *uuid.UUID, environment, projectID string) error {
	meta := &dbmodels.TeamMetadata{
		TeamID: teamID,
		Key:    dbmodels.TeamMetaGoogleProjectID + ":" + environment,
		Value:  &projectID,
	}
	tx := db.Save(meta)
	return tx.Error
}
