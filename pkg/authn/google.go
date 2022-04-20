package authn

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Google struct {
	oauth2.Config

	clientID     string
	clientSecret string
	redirectURL  string

	provider *oidc.Provider
}

func NewGoogle(clientID, clientSecret, redirectURL string) *Google {
	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		panic(err)
	}

	g := &Google{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		provider:     provider,
	}
	g.setupOAuth2()
	return g
}

func (a *Google) setupOAuth2() {
	a.Config = oauth2.Config{
		ClientID:     a.clientID,
		ClientSecret: a.clientSecret,
		Endpoint:     a.provider.Endpoint(),
		RedirectURL:  a.redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
}

func (g *Google) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return g.provider.Verifier(&oidc.Config{ClientID: g.clientID}).Verify(ctx, rawIDToken)
}
