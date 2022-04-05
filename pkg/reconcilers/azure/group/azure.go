package azure_group_reconciler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/registry"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/microsoft"
)

// gitHubReconciler creates teams on GitHub and connects users to them.
type azureReconciler struct {
	logger auditlogger.Logger
	oauth  clientcredentials.Config
}

const (
	Name            = "azure:group"
	OpCreate        = "azure:group:create"
	OpAddMember     = "azure:group:add-member"
	OpAddMembers    = "azure:group:add-members"
	OpDeleteMember  = "azure:group:delete-member"
	OpDeleteMembers = "azure:group:delete-members"
)

func init() {
	registry.Register(Name, NewFromConfig)
}

func New(logger auditlogger.Logger, oauth clientcredentials.Config) *azureReconciler {
	return &azureReconciler{
		logger: logger,
		oauth:  oauth,
	}
}

func NewFromConfig(cfg *config.Config, logger auditlogger.Logger) (reconcilers.Reconciler, error) {
	if !cfg.Azure.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	endpoint := microsoft.AzureADEndpoint(cfg.Azure.TenantID)
	conf := clientcredentials.Config{
		ClientID:     cfg.Azure.ClientID,
		ClientSecret: cfg.Azure.ClientSecret,
		TokenURL:     endpoint.TokenURL,
		AuthStyle:    endpoint.AuthStyle,
		Scopes: []string{
			"https://graph.microsoft.com/.default",
		},
	}

	return New(logger, conf), nil
}

type GroupResponse struct {
	Value []*Group
}

type Group struct {
	ID              string   `json:"id,omitempty"`
	Description     string   `json:"description,omitempty"`
	DisplayName     string   `json:"displayName,omitempty"`
	GroupTypes      []string `json:"groupTypes,omitempty"`
	MailEnabled     bool     `json:"mailEnabled"`
	MailNickname    string   `json:"mailNickname,omitempty"`
	SecurityEnabled bool     `json:"securityEnabled"`
}

func (s *azureReconciler) Reconcile(ctx context.Context, in reconcilers.Input) error {
	grp, err := s.createGroup(ctx, in)
	if err != nil {
		return err
	}

	json.NewEncoder(os.Stdout).Encode(grp)

	return nil
}

func (s *azureReconciler) getGroup(ctx context.Context, slug string) (*Group, error) {
	client := s.oauth.Client(ctx)

	v := &url.Values{}
	v.Add("$filter", fmt.Sprintf("mailNickname eq '%s'", slug))
	u := "https://graph.microsoft.com/v1.0/groups?" + v.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	grp := &GroupResponse{}
	err = dec.Decode(grp)

	if err != nil {
		return nil, err
	}

	switch len(grp.Value) {
	case 0:
		return nil, fmt.Errorf("azure group '%s' does not exist", slug)
	case 1:
		break
	default:
		return nil, fmt.Errorf("ambigious response; more than one search result for azure group '%s'", slug)
	}

	return grp.Value[0], nil
}

func (s *azureReconciler) createGroup(ctx context.Context, in reconcilers.Input) (*Group, error) {
	client := s.oauth.Client(ctx)
	slug := reconcilers.TeamNamePrefix + *in.Team.Slug

	u := "https://graph.microsoft.com/v1.0/groups"

	grp := &Group{
		Description:     *in.Team.Purpose,
		DisplayName:     *in.Team.Name,
		GroupTypes:      nil,
		MailEnabled:     false,
		MailNickname:    slug,
		SecurityEnabled: true,
	}

	payload, err := json.Marshal(grp)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		text, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create azure group '%s': %s: %s", slug, resp.Status, string(text))
	}

	dec := json.NewDecoder(resp.Body)
	grp = &Group{}
	err = dec.Decode(grp)

	if err != nil {
		return nil, err
	}

	if len(grp.ID) == 0 {
		return nil, fmt.Errorf("azure group '%s' created, but no ID returned", slug)
	}

	return grp, nil
}
