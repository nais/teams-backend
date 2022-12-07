package google_gcp_reconciler

import (
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/gcp"
	"github.com/nais/console/pkg/logger"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/iam/v1"
)

type GcpServices struct {
	CloudBillingProjectsService           *cloudbilling.ProjectsService
	CloudResourceManagerProjectsService   *cloudresourcemanager.ProjectsService
	CloudResourceManagerOperationsService *cloudresourcemanager.OperationsService
	IamProjectsServiceAccountsService     *iam.ProjectsServiceAccountsService
}

type googleGcpReconciler struct {
	database       db.Database
	auditLogger    auditlogger.AuditLogger
	clusters       gcp.Clusters
	gcpServices    *GcpServices
	tenantName     string
	domain         string
	cnrmRoleName   string
	billingAccount string
	log            logger.Logger
}
