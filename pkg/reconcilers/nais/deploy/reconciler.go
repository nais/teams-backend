package nais_deploy_reconciler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/registry"
	"gorm.io/gorm"
)

type ProvisionApiKeyRequest struct {
	Team      string
	Rotate    bool
	Timestamp int64
}

// naisDeployReconciler creates teams on GitHub and connects users to them.
type naisDeployReconciler struct {
	client       *http.Client
	logger       auditlogger.Logger
	endpoint     string
	provisionKey []byte
}

const (
	Name              = "nais:deploy"
	OpProvisionApiKey = "nais:deploy:provision-api-key"
)

func init() {
	registry.Register(Name, NewFromConfig)
}

func New(logger auditlogger.Logger, client *http.Client, endpoint string, provisionKey []byte) *naisDeployReconciler {
	return &naisDeployReconciler{
		client:       client,
		logger:       logger,
		endpoint:     endpoint,
		provisionKey: provisionKey,
	}
}

func NewFromConfig(_ *gorm.DB, cfg *config.Config, logger auditlogger.Logger) (reconcilers.Reconciler, error) {
	if !cfg.NaisDeploy.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	provisionKey, err := hex.DecodeString(cfg.NaisDeploy.ProvisionKey)
	if err != nil {
		return nil, err
	}

	return New(logger, http.DefaultClient, cfg.NaisDeploy.Endpoint, provisionKey), nil
}

func (s *naisDeployReconciler) Reconcile(ctx context.Context, in reconcilers.Input) error {
	const signatureHeader = "X-NAIS-Signature"

	payload, err := json.Marshal(&ProvisionApiKeyRequest{
		Rotate:    false,
		Team:      in.Team.Slug.String(),
		Timestamp: time.Now().Unix(),
	})

	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	signature := genMAC(payload, s.provisionKey)
	req.Header.Set(signatureHeader, signature)
	req.Header.Set("content-type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		s.logger.Logf(in, OpProvisionApiKey, "provisioned NAIS deploy API key to team '%s'", in.Team.Slug)
		return nil
	}

	return s.logger.Errorf(in, OpProvisionApiKey, "provision NAIS deploy API key to team '%s': %s", in.Team.Slug, resp.Status)
}

// GenMAC generates the HMAC signature for a message provided the secret key using SHA256
func genMAC(message, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}
