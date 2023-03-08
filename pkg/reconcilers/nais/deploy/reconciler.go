package nais_deploy_reconciler

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/metrics"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

type naisDeployReconciler struct {
	database     db.Database
	client       *http.Client
	auditLogger  auditlogger.AuditLogger
	endpoint     string
	provisionKey []byte
	log          logger.Logger
}

const (
	Name              = sqlc.ReconcilerNameNaisDeploy
	metricsSystemName = "nais-deploy"
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, client *http.Client, endpoint string, provisionKey []byte, log logger.Logger) *naisDeployReconciler {
	return &naisDeployReconciler{
		database:     database,
		client:       client,
		auditLogger:  auditLogger,
		endpoint:     endpoint,
		provisionKey: provisionKey,
		log:          log,
	}
}

func NewFromConfig(_ context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	log = log.WithSystem(string(Name))

	provisionKey, err := hex.DecodeString(cfg.NaisDeploy.ProvisionKey)
	if err != nil {
		return nil, err
	}

	return New(database, auditLogger, http.DefaultClient, cfg.NaisDeploy.Endpoint, provisionKey, log), nil
}

func (r *naisDeployReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *naisDeployReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	payload, err := getProvisionPayload(input.Team.Slug)
	if err != nil {
		return fmt.Errorf("create JSON payload for deploy key API: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request for deploy key API: %w", err)
	}

	signature := genMAC(payload, r.provisionKey)
	request.Header.Set("X-NAIS-Signature", signature)
	request.Header.Set("Content-Type", "application/json")

	response, err := r.client.Do(request)
	metrics.IncExternalHTTPCalls(metricsSystemName, response, err)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusCreated:
		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
		}
		fields := auditlogger.Fields{
			Action:        sqlc.AuditActionNaisDeployProvisionDeployKey,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, r.database, targets, fields, "Provisioned NAIS deploy API key for team %q", input.Team.Slug)

		now := time.Now()
		err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, &reconcilers.NaisDeployKeyState{
			Provisioned: &now,
		})
		if err != nil {
			r.log.WithError(err).Error("persiste reconsiler state")
		}
		return nil
	case http.StatusNoContent:
		return nil
	default:
		return fmt.Errorf("provision NAIS deploy API key for team %q: %s", input.Team.Slug, response.Status)
	}
}

func (r *naisDeployReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	return nil
}
