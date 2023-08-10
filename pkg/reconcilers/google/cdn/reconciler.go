package google_cdn_reconciler

import (
	"context"
	"fmt"

	"google.golang.org/api/iam/v1"

	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/gcp"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/reconcilers"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/nais/teams-backend/pkg/types"
)

const (
	Name              = sqlc.ReconcilerNameGoogleGcpCdn
	metricsSystemName = "cdn"
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, clusters gcp.Clusters, gcpServices *GcpServices, tenantName, domain, cnrmRoleName, billingAccount string, log logger.Logger) *googleCdnReconciler {
	return &googleCdnReconciler{
		database:       database,
		auditLogger:    auditLogger,
		clusters:       clusters,
		gcpServices:    gcpServices,
		domain:         domain,
		cnrmRoleName:   cnrmRoleName,
		billingAccount: billingAccount,
		tenantName:     tenantName,
		log:            log.WithComponent(types.ComponentNameGoogleGcpProject),
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, log logger.Logger) (reconcilers.Reconciler, error) {
	gcpServices, err := createGcpServices(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return New(
		database,
		auditlogger.New(database, types.ComponentNameGoogleGcpProject, log),
		cfg.GCP.Clusters,
		gcpServices,
		cfg.TenantName,
		cfg.TenantDomain,
		cfg.GCP.CnrmRole,
		cfg.GCP.BillingAccount,
		log,
	), nil
}

func (r *googleCdnReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *googleCdnReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	var err error

	const objectViewerRole = "roles/storage.objectViewer" // Define the IAM role for object viewers.
	const objectAdminRole = "roles/storage.objectAdmin"   // Define the IAM role for object admins.

	const bucketNamePrefix = "frontend-plattform" // static?
	const cacheInvalidatorRoleID = "fixme"        // from TF
	const projectID = "fixme"                     // from TF: common project "frontend-plattform"

	cacheInvalidatorRoleName := fmt.Sprintf("projects/%s/roles/%s", projectID, cacheInvalidatorRoleID)

	bucketName := fmt.Sprintf("%s-%s-%s", bucketNamePrefix, r.tenantName, input.Team.Slug) // frontend-plattform-ENV-TEAMSLUG
	bucketOwner := fmt.Sprintf("%s@%s", input.Team.Slug, r.domain)                         // all users in team

	// Create an IAM binding for object viewers.
	objectViewerBinding := &iam.Binding{
		Role:    objectViewerRole,
		Members: []string{"allUsers"},
	}
	_, err = r.gcpServices.IamService.Projects.ServiceAccounts.SetIamPolicy(bucketName, &iam.SetIamPolicyRequest{
		Policy: &iam.Policy{
			Bindings: []*iam.Binding{objectViewerBinding},
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("create IAM binding for object viewers: %v", err)
	}

	// Create an IAM binding for object admins.
	objectAdminBinding := &iam.Binding{
		Role:    objectAdminRole,
		Members: []string{bucketOwner},
	}
	_, err = r.gcpServices.IamService.Projects.ServiceAccounts.SetIamPolicy(bucketName, &iam.SetIamPolicyRequest{
		Policy: &iam.Policy{
			Bindings: []*iam.Binding{objectAdminBinding},
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("create IAM binding for object admins: %v", err)
	}

	// Create an IAM binding for cache invalidators.
	cacheInvalidatorBinding := &iam.Binding{
		Role:    cacheInvalidatorRoleName,
		Members: []string{bucketOwner},
	}
	_, err = r.gcpServices.IamService.Projects.ServiceAccounts.SetIamPolicy(projectID, &iam.SetIamPolicyRequest{
		Policy: &iam.Policy{
			Bindings: []*iam.Binding{cacheInvalidatorBinding},
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("create IAM binding for cache invalidators: %v", err)
	}

	// Create an IAM binding for project viewers.
	projectViewerBinding := &iam.Binding{
		Role:    "roles/viewer",
		Members: []string{bucketOwner},
	}
	_, err = r.gcpServices.IamService.Projects.ServiceAccounts.SetIamPolicy(projectID, &iam.SetIamPolicyRequest{
		Policy: &iam.Policy{
			Bindings: []*iam.Binding{projectViewerBinding},
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("create IAM binding for project viewers: %v", err)
	}

	return nil
}
