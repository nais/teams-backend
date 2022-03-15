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
	"github.com/nais/console/pkg/reconcilers"
)

type provisionApiKeyRequest struct {
	Team      string
	Rotate    bool
	Timestamp int64
}

// naisDeployReconciler creates teams on GitHub and connects users to them.
type naisDeployReconciler struct {
	logger       auditlogger.Logger
	endpoint     string
	provisionKey []byte
}

func New(logger auditlogger.Logger, endpoint string, provisionKey []byte) *naisDeployReconciler {
	return &naisDeployReconciler{
		logger:       logger,
		endpoint:     endpoint,
		provisionKey: provisionKey,
	}
}

const (
	Name              = "nais:deploy"
	OpProvisionApiKey = "nais:deploy:provision-api-key"
)

func (s *naisDeployReconciler) Name() string {
	return Name
}

func (s *naisDeployReconciler) Reconcile(ctx context.Context, in reconcilers.Input) error {
	const signatureHeader = "X-NAIS-Signature"

	payload, err := json.Marshal(&provisionApiKeyRequest{
		Rotate:    false,
		Team:      *in.Team.Slug,
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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusCreated {
		s.logger.Logf(in, OpProvisionApiKey, "provisioned NAIS deploy API key to team '%s'", *in.Team.Slug)
		return nil
	}

	return s.logger.Errorf(in, OpProvisionApiKey, "provision NAIS deploy API key to team '%s': %s", *in.Team.Slug, resp.Status)
}

// GenMAC generates the HMAC signature for a message provided the secret key using SHA256
func genMAC(message, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}
