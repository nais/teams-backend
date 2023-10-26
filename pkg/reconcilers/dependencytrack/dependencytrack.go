package dependencytrack_reconciler

import (
	"net/http"

	dependencytrack "github.com/nais/dependencytrack/pkg/client"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/metrics"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/sirupsen/logrus"
)

type DpTrack struct {
	Endpoint string
	Client   dependencytrack.Client
}

func NewDpTrackWithClient(endpoint string, client dependencytrack.Client, log logger.Logger) DpTrack {
	dp := DpTrack{
		Endpoint: endpoint,
		Client:   client,
	}
	dependencytrack.WithLogger(log.WithFields(logrus.Fields{
		"instance": endpoint,
	}))
	dependencytrack.WithResponseCallback(incExternalHttpCalls)
	return dp
}

func newDpTrack(endpoint, username, password string, log logger.Logger) DpTrack {
	return NewDpTrackWithClient(endpoint, dependencytrack.New(endpoint, username, password), log)
}

func incExternalHttpCalls(resp *http.Response, err error) {
	metrics.IncExternalHTTPCalls(string(sqlc.ReconcilerNameNaisDependencytrack), resp, err)
}
