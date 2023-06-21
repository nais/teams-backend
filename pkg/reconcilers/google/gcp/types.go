package google_gcp_reconciler

import (
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/gcp"
	"github.com/nais/teams-backend/pkg/logger"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/serviceusage/v1"
)

type GcpServices struct {
	CloudBillingProjectsService           *cloudbilling.ProjectsService
	CloudResourceManagerProjectsService   *cloudresourcemanager.ProjectsService
	CloudResourceManagerOperationsService *cloudresourcemanager.OperationsService
	IamProjectsServiceAccountsService     *iam.ProjectsServiceAccountsService
	ServiceUsageService                   *serviceusage.ServicesService
	ServiceUsageOperationsService         *serviceusage.OperationsService
	FirewallService                       *compute.FirewallsService
	ComputeGlobalOperationsService        *compute.GlobalOperationsService
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
