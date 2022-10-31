package google_gcp_reconciler

import (
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/iam/v1"
)

type Cluster struct {
	TeamFolderID int64  `json:"teams_folder_id"`
	ProjectID    string `json:"project_id"`
}

type ClusterInfo map[string]Cluster

type GcpServices struct {
	CloudBillingProjectsService           *cloudbilling.ProjectsService
	CloudResourceManagerProjectsService   *cloudresourcemanager.ProjectsService
	CloudResourceManagerOperationsService *cloudresourcemanager.OperationsService
	IamProjectsServiceAccountsService     *iam.ProjectsServiceAccountsService
}

type googleGcpReconciler struct {
	database       db.Database
	auditLogger    auditlogger.AuditLogger
	clusters       ClusterInfo
	gcpServices    *GcpServices
	tenantName     string
	domain         string
	cnrmRoleName   string
	billingAccount string
}
