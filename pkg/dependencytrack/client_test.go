package dependencytrack

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestClient_Token(t *testing.T) {
	log, err := logger.GetLogger("text", "debug")
	assert.NoError(t, err)

	t.Run("login success, new token retrieved", func(t *testing.T) {
		httpClient := test.NewTestHttpClient(
			successfulLogin(t),
		)
		c := NewClient("http://localhost/api/v1", "username", "password", httpClient, log)

		token, err := c.(*client).token(context.Background())
		assert.NoError(t, err)
		fmt.Printf("token: %s\n", token)
	})

	t.Run("login failed, not found", func(t *testing.T) {
		httpClient := test.NewTestHttpClient(
			func(req *http.Request) *http.Response {
				assert.Equal(t, "POST", req.Method)

				return test.Response("404", "Not found")
			},
		)
		c := NewClient("http://localhost", "username", "password", httpClient, log)

		token, err := c.(*client).token(context.Background())
		assert.Error(t, err)
		fmt.Printf("token: %s\n", token)
	})

	t.Run("login failed, invalid credentials", func(t *testing.T) {
		httpClient := test.NewTestHttpClient(
			func(req *http.Request) *http.Response {
				assert.Equal(t, "POST", req.Method)

				return test.Response("401", "INVALID_CREDENTIALS")
			},
		)
		c := NewClient("http://localhost", "username", "password", httpClient, log)

		token, err := c.(*client).token(context.Background())
		assert.Error(t, err)
		fmt.Printf("token: %s\n", token)
	})
}

func TestClient_CreateTeam(t *testing.T) {
	log, err := logger.GetLogger("text", "debug")
	assert.NoError(t, err)

	t.Run("CreateTeam success", func(t *testing.T) {
		httpClient := test.NewTestHttpClient(
			successfulLogin(t),
			func(req *http.Request) *http.Response {
				assert.Equal(t, "PUT", req.Method)
				assert.Equal(t, "/api/v1/team", req.URL.Path)

				return test.Response("200", "{ \"name\": \"test\", \"uuid\": \"1234\" }")
			},
		)

		c := NewClient("http://localhost/api/v1", "username", "password", httpClient, log)
		team, err := c.CreateTeam(context.Background(), "test", []Permission{})
		assert.NoError(t, err)
		assert.Equal(t, "test", team.Name)
	})
}

func successfulLogin(t *testing.T) func(req *http.Request) *http.Response {
	return func(req *http.Request) *http.Response {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "/api/v1/user/login", req.URL.Path)

		return test.Response("200", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")
	}
}
