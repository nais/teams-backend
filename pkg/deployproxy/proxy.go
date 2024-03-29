package deployproxy

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nais/teams-backend/pkg/graph/apierror"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/metrics"
	"github.com/nais/teams-backend/pkg/slug"
)

const metricSystemName = "deploy-proxy"

type Proxy interface {
	GetApiKey(ctx context.Context, slug slug.Slug) (string, error)
}

type deploy struct {
	client       *http.Client
	endpoint     string
	provisionKey []byte
	log          logger.Logger
}

type response struct {
	Message string   `json:"message,omitempty"`
	ApiKeys []string `json:"apiKeys,omitempty"`
}

func NewProxy(endpoint string, provisionString string, log logger.Logger) (Proxy, error) {
	provisionKey, err := hex.DecodeString(provisionString)
	if err != nil {
		return nil, err
	}

	return &deploy{
		client:       http.DefaultClient,
		endpoint:     endpoint,
		provisionKey: provisionKey,
		log:          log,
	}, nil
}

func (d *deploy) GetApiKey(ctx context.Context, slug slug.Slug) (string, error) {
	payload, err := getDeployKeyPayload(slug)
	if err != nil {
		return "", fmt.Errorf("create JSON payload for deploy key API: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create request for deploy key API: %w", err)
	}

	signature := genMAC(payload, d.provisionKey)
	request.Header.Set("X-NAIS-Signature", signature)
	request.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(request)
	metrics.IncExternalHTTPCalls(metricSystemName, resp, err)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	response := response{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return "", fmt.Errorf("unable to unmarsal reply from deploy API: %v", err)
	}

	if len(response.ApiKeys) < 1 {
		return "", apierror.Errorf("team %q has no deploy keys", slug)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		d.log.Debugf("Returning first of %d deploy keys for team %q", len(response.ApiKeys), slug)
		return response.ApiKeys[0], nil
	case http.StatusNotFound:
		return "", apierror.Errorf("team %q is not provisioned in NAIS deploy", slug)
	default:
		return "", fmt.Errorf("failed to get deploy key for team %q: %s", slug, response.Message)
	}
}

type DeployKeyRequest struct {
	Team      string
	Timestamp int64
}

// getDepoyKeyPayload get a payload for the NAIS deploy deploy key request
func getDeployKeyPayload(slug slug.Slug) ([]byte, error) {
	return json.Marshal(&DeployKeyRequest{
		Team:      string(slug),
		Timestamp: time.Now().Unix(),
	})
}

// genMAC generates the HMAC signature for a message provided the secret key using SHA256
func genMAC(message, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}
