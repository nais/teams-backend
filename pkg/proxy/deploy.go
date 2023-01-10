package proxy

import (
	"fmt"
	"github.com/nais/console/pkg/slug"
	"net/http"
)

type Deploy interface {
	GetApiKey(slug slug.Slug) (string, error)
}

type deploy struct {
	client *http.Client
}

func NewDeploy() Deploy {
	return &deploy{
		client: http.DefaultClient,
	}
}

func (d *deploy) GetApiKey(slug slug.Slug) (string, error) {
	// TODO: implement me
	return fmt.Sprintf("TODO implement me. slug: %s", slug), nil
}
