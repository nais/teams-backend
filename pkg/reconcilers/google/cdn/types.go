package google_cdn_reconciler

import (
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/storage/v1"

	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/gcp"
	"github.com/nais/teams-backend/pkg/logger"
)

type GcpServices struct {
	BucketsService *storage.BucketsService
	IamService     *iam.Service
}

type googleCdnReconciler struct {
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
