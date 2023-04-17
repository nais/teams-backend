package dependencytrack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nais/console/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

const metricsSystemName = "dependencytrack"

type Client struct {
	baseUrl     string
	username    string
	password    string
	accessToken string
	httpClient  *http.Client
}

type Permission string

const ViewPortfolioPermission = Permission("VIEW_PORTFOLIO")

type Team struct {
	Uuid      string `json:"uuid,omitempty"`
	Name      string `json:"name,omitempty"`
	OidcUsers []User `json:"oidcUsers,omitempty"`
}

func NewClient(baseUrl string, username string, password string) *Client {
	return &Client{
		baseUrl:    baseUrl,
		username:   username,
		password:   password,
		httpClient: http.DefaultClient,
	}
}

type User struct {
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

type RequestError struct {
	StatusCode int
	Err        error
}

// TODO: check if team exists - name is not unique
func (c *Client) CreateTeam(ctx context.Context, teamName string, permissions []Permission) (*Team, error) {
	body, _ := json.Marshal(&Team{
		Name: teamName,
	})

	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}

	t := &Team{}
	b, err := c.sendRequest(ctx, http.MethodPut, c.baseUrl+"/team", map[string][]string{
		"Content-Type":  {"application/json"},
		"Accept":        {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}, body)

	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, t); err != nil {
		return nil, err
	}

	for _, p := range permissions {
		if _, err := c.sendRequest(ctx, http.MethodPost, c.baseUrl+"/permission/"+string(p)+"/team/"+t.Uuid, map[string][]string{
			"Accept":        {"application/json"},
			"Authorization": {fmt.Sprintf("Bearer %s", token)},
		}, nil); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (c *Client) GetTeams(ctx context.Context) ([]Team, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}

	b, err := c.sendRequest(ctx, http.MethodGet, c.baseUrl+"/team", map[string][]string{
		"Content-Type":  {"application/json"},
		"Accept":        {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}, nil)

	if err != nil {
		return nil, err
	}

	var teams []Team
	if err := json.Unmarshal(b, &teams); err != nil {
		return nil, err
	}
	return teams, nil
}

func GetTeamUuid(teams []Team, name string) string {
	for _, t := range teams {
		if t.Name == name {
			return t.Uuid
		}
	}
	return ""
}

func (c *Client) CreateUser(ctx context.Context, email string) error {
	body, err := json.Marshal(map[string]string{
		"username": email,
		"email":    email,
	})

	token, err := c.token(ctx)
	if err != nil {
		return err
	}

	_, err = c.sendRequest(ctx, http.MethodPut, c.baseUrl+"/user/oidc", map[string][]string{
		"Content-Type":  {"application/json"},
		"Accept":        {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}, body)
	if err != nil {
		e, ok := err.(*RequestError)
		if !ok {
			return fmt.Errorf("create user: %w", err)
		}
		if e.StatusCode == http.StatusConflict {
			log.Infof("user %s already exists", email)
			return nil
		}
		return fmt.Errorf("create user: %w", err)

	}
	return nil
}

func (c *Client) AddToTeam(ctx context.Context, username string, uuid string) error {
	token, err := c.token(ctx)
	if err != nil {
		return fmt.Errorf("getting Token: %w", err)
	}

	_, err = c.sendRequest(ctx, http.MethodPost, c.baseUrl+"/user/"+username+"/membership", map[string][]string{
		"Content-Type":  {"application/json"},
		"Accept":        {"application/json"},
		"Authorization": {"Bearer " + token},
	}, []byte(`{"uuid": "`+uuid+`"}`))

	if err != nil {
		e, ok := err.(*RequestError)
		if !ok {
			return fmt.Errorf("adding user to team: %w", err)
		}
		if e.StatusCode == http.StatusNotModified {
			log.Infof("user %s already in team", username)
			return nil
		}
		return fmt.Errorf("adding user to team: %w", err)
	}
	return nil
}

func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting Token: %w", err)
	}
	b, err := c.sendRequest(ctx, http.MethodGet, c.baseUrl+"/user/oidc", map[string][]string{
		"Accept":        {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}
	var users []User
	if err := json.Unmarshal(b, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (c *Client) DeleteUser(ctx context.Context, email string) error {
	body, err := json.Marshal(map[string]string{
		"username": email,
		"email":    email,
	})
	token, err := c.token(ctx)
	if err != nil {
		return fmt.Errorf("getting Token: %w", err)
	}
	_, err = c.sendRequest(ctx, http.MethodDelete, c.baseUrl+"/user/oidc", map[string][]string{
		"Content-Type":  {"application/json"},
		"Accept":        {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}, body)
	if err != nil {
		e, ok := err.(*RequestError)
		if !ok {
			return fmt.Errorf("deleting user: %w", err)
		}
		if e.StatusCode == http.StatusNotFound {
			log.Infof("user %s does not exist", email)
			return nil
		}
		return fmt.Errorf("deleting user: %w", err)
	}
	return nil
}

func (c *Client) DeleteTeam(ctx context.Context, team string) error {
	teams, err := c.GetTeams(ctx)
	teamUuid := GetTeamUuid(teams, team)
	body, err := json.Marshal(map[string]string{
		"uuid": teamUuid,
	})
	token, err := c.token(ctx)
	if err != nil {
		return fmt.Errorf("getting Token: %w", err)
	}
	_, err = c.sendRequest(ctx, http.MethodDelete, c.baseUrl+"/team", map[string][]string{
		"Content-Type":  {"application/json"},
		"Accept":        {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}, body)
	if err != nil {
		e, ok := err.(*RequestError)
		if !ok {
			return fmt.Errorf("deleting team: %w", err)
		}
		if e.StatusCode == http.StatusNotFound {
			log.Infof("team %s does not exist", team)
			return nil
		}
		return fmt.Errorf("deleting team: %w", err)
	}
	return nil
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("status %d: err %v", r.StatusCode, r.Err)
}

func (r *RequestError) AlreadyExists() bool {
	return r.StatusCode == http.StatusConflict
}

func (c *Client) sendRequest(ctx context.Context, httpMethod string, url string, headers map[string][]string, body []byte) ([]byte, error) {
	fmt.Printf("Sending request to %s\n", url)
	req, err := http.NewRequestWithContext(ctx, httpMethod, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header = headers

	resp, err := c.httpClient.Do(req)
	metrics.IncExternalHTTPCalls(metricsSystemName, resp, err)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}
		return nil, fail(resp.StatusCode, fmt.Errorf("%s\n", string(b)))
	}
	resBody, err := io.ReadAll(resp.Body)
	return resBody, err
}

func fail(status int, err error) *RequestError {
	return &RequestError{
		StatusCode: status,
		Err:        err,
	}
}
