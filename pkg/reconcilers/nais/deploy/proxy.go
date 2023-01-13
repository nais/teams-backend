package nais_deploy_reconciler

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/nais/console/pkg/graph/apierror"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/metrics"
	"github.com/nais/console/pkg/slug"
	"io"
	"net/http"
)

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
	metrics.IncExternalHTTPCalls(metricsSystemName, resp, err)
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
