package dependencytrack_reconciler

import "github.com/nais/dependencytrack/pkg/client"

type Client interface {
	client.Client
}
