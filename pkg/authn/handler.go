package authn

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	log "github.com/sirupsen/logrus"
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

type Handler struct {
	database     db.Database
	oauth2Config OAuth2
	frontendURL  url.URL
}

func New(oauth2Config OAuth2, database db.Database, frontendURL url.URL) *Handler {
	return &Handler{
		database:     database,
		oauth2Config: oauth2Config,
		frontendURL:  frontendURL,
	}
}

type claims struct {
	Email string
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	http.SetCookie(w, &http.Cookie{
		Name:     RedirectURICookie,
		Value:    redirectURI,
		Path:     "/",
		Expires:  time.Now().Add(30 * time.Minute),
		Secure:   true,
		HttpOnly: true,
	})

	oauthState := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:     OAuthStateCookie,
		Value:    oauthState,
		Path:     "/",
		Expires:  time.Now().Add(30 * time.Minute),
		Secure:   true,
		HttpOnly: true,
	})
	consentUrl := h.oauth2Config.AuthCodeURL(oauthState, oauth2.SetAuthURLParam("prompt", "select_account"))
	http.Redirect(w, r, consentUrl, http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	frontendURL := h.frontendURL

	redirectURI, err := r.Cookie(RedirectURICookie)
	if err == nil {
		frontendURL.Path = strings.TrimPrefix(redirectURI.Value, "/")
	}

	deleteCookie(w, RedirectURICookie)
	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		log.Error("Missing code query parameter")
		http.Redirect(w, r, frontendURL.String()+"?error=unauthenticated", http.StatusFound)
		return
	}

	oauthCookie, err := r.Cookie(OAuthStateCookie)
	if err != nil {
		log.WithError(err).Error("Missing oauth state cookie")
		http.Redirect(w, r, frontendURL.String()+"?error=invalid-state", http.StatusFound)
		return
	}

	deleteCookie(w, OAuthStateCookie)
	state := r.URL.Query().Get("state")
	if state != oauthCookie.Value {
		log.Error("Incoming state does not match local state")
		http.Redirect(w, r, frontendURL.String()+"?error=invalid-state", http.StatusFound)
		return
	}

	tokens, err := h.oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		log.WithError(err).Error("Exchanging authorization code for tokens")
		http.Redirect(w, r, frontendURL.String()+"?error=unauthenticated", http.StatusFound)
		return
	}

	rawIDToken, ok := tokens.Extra(IDTokenKey).(string)
	if !ok {
		log.Error("Missing id_token")
		http.Redirect(w, r, frontendURL.String()+"?error=unauthenticated", http.StatusFound)
		return
	}

	idToken, err := h.oauth2Config.Verify(r.Context(), rawIDToken)
	if err != nil {
		log.WithError(err).Error("Invalid id_token")
		http.Redirect(w, r, frontendURL.String()+"?error=unauthenticated", http.StatusFound)
		return
	}

	claims := &claims{}
	if err := idToken.Claims(claims); err != nil {
		log.WithError(err).Error("Unable to parse claims")
		http.Redirect(w, r, frontendURL.String()+"?error=unauthenticated", http.StatusFound)
		return
	}

	user, err := h.database.GetUserByEmail(r.Context(), claims.Email)
	if err != nil {
		log.WithError(err).Errorf("User with email %q does not exist in the Console database", claims.Email)
		http.Redirect(w, r, frontendURL.String()+"?error=unknown-user", http.StatusFound)
		return
	}

	session, err := h.database.CreateSession(r.Context(), user.ID)
	if err != nil {
		log.WithError(err).Errorf("Unable to create session")
		http.Redirect(w, r, frontendURL.String()+"?error=unable-to-create-session", http.StatusFound)
		return
	}

	SetSessionCookie(w, session)
	http.Redirect(w, r, frontendURL.String(), http.StatusFound)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	redirectUrl := "/"
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		log.WithError(err).Error("Unable to logout session")
		http.Redirect(w, r, redirectUrl, http.StatusFound)
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		log.WithError(err).Error("Unable to parse cookie value as UUID")
		http.Redirect(w, r, redirectUrl, http.StatusFound)
		return
	}

	err = h.database.DeleteSession(r.Context(), sessionID)
	if err != nil {
		log.WithError(err).Error("Unable to delete session from database")
		http.Redirect(w, r, redirectUrl, http.StatusFound)
		return
	}

	DeleteSessionCookie(w)
	http.Redirect(w, r, redirectUrl, http.StatusFound)
}

func SetSessionCookie(w http.ResponseWriter, session *db.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    session.ID.String(),
		Path:     "/",
		Expires:  session.Expires,
		Secure:   true,
		HttpOnly: true,
	})
}

func deleteCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		Secure:   true,
		HttpOnly: true,
	})
}

func DeleteSessionCookie(w http.ResponseWriter) {
	deleteCookie(w, SessionCookieName)
}
