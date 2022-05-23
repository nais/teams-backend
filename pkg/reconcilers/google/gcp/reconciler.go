package google_gcp_reconciler

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/registry"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type gcpReconciler struct {
	config           *jwt.Config
	domain           string
	logger           auditlogger.Logger
	projectParentIDs map[string]string
}

const (
	Name                    = "google:workspace-admin"
	OpCreate                = "google:workspace-admin:create"
	OpAddMember             = "google:workspace-admin:add-member"
	OpAddMembers            = "google:workspace-admin:add-members"
	OpDeleteMember          = "google:workspace-admin:delete-member"
	OpDeleteMembers         = "google:workspace-admin:delete-members"
	OpAddToGKESecurityGroup = "google:workspace-admin:add-to-gke-security-group"
)

func init() {
	registry.Register(Name, NewFromConfig)
}

func New(logger auditlogger.Logger, domain string, config *jwt.Config, projectParentIDs map[string]string) *gcpReconciler {
	return &gcpReconciler{
		logger:           logger,
		domain:           domain,
		config:           config,
		projectParentIDs: projectParentIDs,
	}
}

func NewFromConfig(cfg *config.Config, logger auditlogger.Logger) (reconcilers.Reconciler, error) {
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

	return New(logger, cfg.GCP.Domain, cf, cfg.GCP.ProjectParentIDs), nil
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

		err = s.CreatePermissions(svc, in, proj.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *gcpReconciler) CreateProject(svc *cloudresourcemanager.Service, in reconcilers.Input, environment, parentID string) (*cloudresourcemanager.Project, error) {
	// TODO: what if our deterministic "globally unique" project ID is taken?

	projectID := CreateProjectID(s.domain, environment, *in.Team.Slug)

	proj := &cloudresourcemanager.Project{
		Parent:    parentID,
		ProjectId: projectID,
	}

	oper, err := svc.Projects.Create(proj).Do()

	switch typedError := err.(type) {
	case *googleapi.Error:
		if typedError.Code != http.StatusConflict {
			return nil, err
		}
		// conflict may be due to
		// 1) already created by us in this folder, or
		// 2) someone else owns this project
		log.Warnf("project creation conflict")

		query, err := svc.Projects.Search().Query("id:" + projectID).Do()
		if err != nil {
			return nil, err
		}

		if len(query.Projects) == 0 {
			return nil, fmt.Errorf("FIXME: project ID is taken")
		}

		for _, proj = range query.Projects {
			if proj.ProjectId == projectID {
				return proj, nil
			}
		}

		return nil, fmt.Errorf("BUG: search results for project ID returned project without correct ID")

	case nil:
		for !oper.Done {
			var err error
			oper, err = svc.Operations.Get(oper.Name).Do()
			if err != nil {
				return nil, err
			}
		}

		if oper.Error != nil {
			return nil, errors.New(oper.Error.Message)
		}

	default:
		return nil, err
	}

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

	return err
}
