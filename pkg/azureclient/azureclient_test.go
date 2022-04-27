package azureclient

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func Test_client_GetUser(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/users/user@example.com", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`{
				"mail": "user@example.com",
				"id": "some-id"
			}`)),
			Header: make(http.Header),
		}
	})

	client := New(httpClient)
	member, err := client.GetUser(context.TODO(), "user@example.com")

	assert.Equal(t, "user@example.com", member.Mail)
	assert.Equal(t, "some-id", member.ID)
	assert.NoError(t, err)
}

func Test_client_GetUserThatDoesNotExist(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/users/user@example.com", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return &http.Response{
			Status:     "404 Not Found",
			StatusCode: 404,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"error": {"message": "user does not exist"}}`)),
			Header:     make(http.Header),
		}
	})

	client := New(httpClient)
	member, err := client.GetUser(context.TODO(), "user@example.com")

	assert.Nil(t, member)
	assert.EqualError(t, err, `404 Not Found: {"error": {"message": "user does not exist"}}`)
}

func Test_client_GetUserWithInvalidApiResponse(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/users/user@example.com", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`some string`)),
			Header:     make(http.Header),
		}
	})

	client := New(httpClient)
	member, err := client.GetUser(context.TODO(), "user@example.com")

	assert.Nil(t, member)
	assert.EqualError(t, err, "invalid character 's' looking for beginning of value")
}

func Test_client_GetGroup(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/groups?%24filter=mailNickname+eq+%27slug%27", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"value": [{"id":"group-id"}]}`)),
			Header:     make(http.Header),
		}
	})

	client := New(httpClient)
	group, err := client.GetGroup(context.TODO(), "slug")

	assert.NoError(t, err)
	assert.Equal(t, "group-id", group.ID)
}

func Test_client_GetGroupThatDoesNotExist(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/groups?%24filter=mailNickname+eq+%27slug%27", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"value": []}`)),
			Header:     make(http.Header),
		}
	})

	client := New(httpClient)
	group, err := client.GetGroup(context.TODO(), "slug")

	assert.Nil(t, group)
	assert.EqualError(t, err, "azure group 'slug' does not exist")
}

func Test_client_GetGroupWithAmbiguousResult(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/groups?%24filter=mailNickname+eq+%27slug%27", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"value": [{"id":"group-id"}, {"id":"group-id-2"}]}`)),
			Header:     make(http.Header),
		}
	})

	client := New(httpClient)
	group, err := client.GetGroup(context.TODO(), "slug")

	assert.Nil(t, group)
	assert.EqualError(t, err, "ambiguous response; more than one search result for azure group 'slug'")
}
