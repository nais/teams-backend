package authn

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/sqlc"
	"golang.org/x/oauth2"
)

const (
	RedirectURICookie = "redirecturi"
	OAuthStateCookie  = "oauthstate"
	SessionCookieName = "console_session_id"
	IDTokenKey        = "id_token"
)

type OAuth2 interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

type Handler interface {
	Login(w http.ResponseWriter, r *http.Request)
	Callback(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	SetSessionCookie(w http.ResponseWriter, session *db.Session)
	DeleteCookie(w http.ResponseWriter, name string)
}

type handler struct {
	database     db.Database
	oauth2Config OAuth2
	frontendURL  url.URL
	secureCookie bool
	log          logger.Logger
}

func New(oauth2Config OAuth2, database db.Database, frontendURL url.URL, log logger.Logger) Handler {
	return &handler{
		database:     database,
		oauth2Config: oauth2Config,
		frontendURL:  frontendURL,
		secureCookie: shouldUseSecureCookies(frontendURL),
		log:          log.WithSystem(string(sqlc.SystemNameAuthn)),
	}
}

type claims struct {
	Email string
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	http.SetCookie(w, &http.Cookie{
		Name:     RedirectURICookie,
		Value:    redirectURI,
		Path:     "/",
		Expires:  time.Now().Add(30 * time.Minute),
		Secure:   h.secureCookie,
		HttpOnly: true,
	})

	oauthState := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:     OAuthStateCookie,
		Value:    oauthState,
		Path:     "/",
		Expires:  time.Now().Add(30 * time.Minute),
		Secure:   h.secureCookie,
		HttpOnly: true,
	})
	consentUrl := h.oauth2Config.AuthCodeURL(oauthState, oauth2.SetAuthURLParam("prompt", "select_account"))
	http.Redirect(w, r, consentUrl, http.StatusFound)
}

func (h *handler) Callback(w http.ResponseWriter, r *http.Request) {
	frontendURL := h.frontendURL

	redirectURI, err := r.Cookie(RedirectURICookie)
	if err == nil {
		frontendURL.Path = strings.TrimPrefix(redirectURI.Value, "/")
	}

	h.DeleteCookie(w, RedirectURICookie)
	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		h.log.WithError(fmt.Errorf("missing query parameter")).Error("check code param")
		http.Redirect(w, r, frontendURL.String()+"?error=unauthenticated", http.StatusFound)
		return
	}

	oauthCookie, err := r.Cookie(OAuthStateCookie)
	if err != nil {
		h.log.WithError(err).Error("missing oauth state cookie")
		http.Redirect(w, r, frontendURL.String()+"?error=invalid-state", http.StatusFound)
		return
	}

	h.DeleteCookie(w, OAuthStateCookie)
	state := r.URL.Query().Get("state")
	if state != oauthCookie.Value {
		h.log.WithError(fmt.Errorf("state mismatch")).Error("check incoming state matches local state")
		http.Redirect(w, r, frontendURL.String()+"?error=invalid-state", http.StatusFound)
		return
	}

	tokens, err := h.oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		h.log.WithError(err).Error("exchanging authorization code for tokens")
		http.Redirect(w, r, frontendURL.String()+"?error=unauthenticated", http.StatusFound)
		return
	}

	rawIDToken, ok := tokens.Extra(IDTokenKey).(string)
	if !ok {
		h.log.WithError(fmt.Errorf("missing id_token")).Error("id token presence")
		http.Redirect(w, r, frontendURL.String()+"?error=unauthenticated", http.StatusFound)
		return
	}

	idToken, err := h.oauth2Config.Verify(r.Context(), rawIDToken)
	if err != nil {
		h.log.WithError(err).Error("verify id_token")
		http.Redirect(w, r, frontendURL.String()+"?error=unauthenticated", http.StatusFound)
		return
	}

	claims := &claims{}
	if err := idToken.Claims(claims); err != nil {
		h.log.WithError(err).Error("parse claims")
		http.Redirect(w, r, frontendURL.String()+"?error=unauthenticated", http.StatusFound)
		return
	}

	user, err := h.database.GetUserByEmail(r.Context(), claims.Email)
	if err != nil {
		h.log.WithError(err).Errorf("get user (%s) from db", claims.Email)
		http.Redirect(w, r, frontendURL.String()+"?error=unknown-user", http.StatusFound)
		return
	}

	session, err := h.database.CreateSession(r.Context(), user.ID)
	if err != nil {
		h.log.WithError(err).Error("create session")
		http.Redirect(w, r, frontendURL.String()+"?error=unable-to-create-session", http.StatusFound)
		return
	}

	h.SetSessionCookie(w, session)
	http.Redirect(w, r, frontendURL.String(), http.StatusFound)
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	redirectUrl := "/"
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		h.log.WithError(err).Error("logout session")
		http.Redirect(w, r, redirectUrl, http.StatusFound)
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		h.log.WithError(err).Error("parse cookie value as UUID")
		http.Redirect(w, r, redirectUrl, http.StatusFound)
		return
	}

	err = h.database.DeleteSession(r.Context(), sessionID)
	if err != nil {
		h.log.WithError(err).Error("delete session from database")
		http.Redirect(w, r, redirectUrl, http.StatusFound)
		return
	}

	http.Redirect(w, r, redirectUrl, http.StatusFound)
}

func (h *handler) SetSessionCookie(w http.ResponseWriter, session *db.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    session.ID.String(),
		Path:     "/",
		Expires:  session.Expires,
		Secure:   h.secureCookie,
		HttpOnly: true,
	})
}

func (h *handler) DeleteCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		Secure:   h.secureCookie,
		HttpOnly: true,
	})
}

func shouldUseSecureCookies(frontendURL url.URL) bool {
	return frontendURL.Host != "teams.local.nais.io"
}
