package azureclient

import (
	"bytes"
	"context"
	helpers "github.com/nais/console/pkg/console"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

type RoundTripFunc func(req *http.Request) *http.Response

type RoundTripper struct {
	funcs          []RoundTripFunc
	requestCounter int
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	handler := r.funcs[r.requestCounter]
	r.requestCounter++
	return handler(req), nil
}

func NewTestClient(funcs ...RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: &RoundTripper{
			funcs: funcs,
		},
	}
}

func Test_GetUser(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/users/user@example.com", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return response("200 OK", `{
			"mail": "user@example.com",
			"id": "some-id"
		}`)
	})

	client := New(httpClient)
	member, err := client.GetUser(context.TODO(), "user@example.com")

	assert.Equal(t, "user@example.com", member.Mail)
	assert.Equal(t, "some-id", member.ID)
	assert.NoError(t, err)
}

func Test_GetUserThatDoesNotExist(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/users/user@example.com", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return response("404 Not Found", `{"error": {"message": "user does not exist"}}`)
	})

	client := New(httpClient)
	member, err := client.GetUser(context.TODO(), "user@example.com")

	assert.Nil(t, member)
	assert.EqualError(t, err, `404 Not Found: {"error": {"message": "user does not exist"}}`)
}

func Test_GetUserWithInvalidApiResponse(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/users/user@example.com", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return response("200 OK", "some string")
	})

	client := New(httpClient)
	member, err := client.GetUser(context.TODO(), "user@example.com")

	assert.Nil(t, member)
	assert.EqualError(t, err, "invalid character 's' looking for beginning of value")
}

func Test_GetGroup(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/groups?%24filter=mailNickname+eq+%27slug%27", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return response("200 OK", `{"value": [{"id":"group-id"}]}`)
	})

	client := New(httpClient)
	group, err := client.GetGroupByMailNickName(context.TODO(), "slug")

	assert.NoError(t, err)
	assert.Equal(t, "group-id", group.ID)
}

func Test_GetGroupThatDoesNotExist(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/groups?%24filter=mailNickname+eq+%27slug%27", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return response("200 OK", `{"value": []}`)
	})

	client := New(httpClient)
	group, err := client.GetGroupByMailNickName(context.TODO(), "slug")

	assert.Nil(t, group)
	assert.EqualError(t, err, "azure group 'slug' does not exist")
}

func Test_GetGroupWithAmbiguousResult(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/groups?%24filter=mailNickname+eq+%27slug%27", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)

		return response("200 OK", `{"value": [{"id":"group-id"}, {"id":"group-id-2"}]}`)
	})

	client := New(httpClient)
	group, err := client.GetGroupByMailNickName(context.TODO(), "slug")

	assert.Nil(t, group)
	assert.EqualError(t, err, "ambiguous response; more than one search result for azure group 'slug'")
}

func Test_CreateGroup(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/groups", req.URL.String())
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "application/json", req.Header.Get("content-type"))

		return response("201 Created", `{
			"id":"some-id",
			"description":"description",
			"displayName": "name",
			"mailNickname": "mail"
		}`)
	})

	client := New(httpClient)
	input := &Group{
		Description:  "description",
		DisplayName:  "name",
		MailNickname: "mail",
	}
	expectedOutput := input
	expectedOutput.ID = "some-id"

	group, err := client.CreateGroup(context.TODO(), input)

	assert.Equal(t, expectedOutput, group)
	assert.NoError(t, err)
}

func Test_CreateGroupWithInvalidStatus(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/groups", req.URL.String())
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "application/json", req.Header.Get("content-type"))

		return response("400 Bad Request", `{"error": {"message":"some error"}}`)
	})

	client := New(httpClient)

	group, err := client.CreateGroup(context.TODO(), &Group{
		Description:  "description",
		DisplayName:  "name",
		MailNickname: "mail",
	})

	assert.Nil(t, group)
	assert.EqualError(t, err, `create azure group 'mail': 400 Bad Request: {"error": {"message":"some error"}}`)
}

func Test_CreateGroupWithInvalidResponse(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/groups", req.URL.String())
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "application/json", req.Header.Get("content-type"))

		return response("201 Created", "response body")
	})

	client := New(httpClient)

	group, err := client.CreateGroup(context.TODO(), &Group{
		Description:  "description",
		DisplayName:  "name",
		MailNickname: "mail",
	})

	assert.Nil(t, group)
	assert.EqualError(t, err, "invalid character 'r' looking for beginning of value")
}

func Test_CreateGroupWithIncompleteResponse(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://graph.microsoft.com/v1.0/groups", req.URL.String())
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "application/json", req.Header.Get("content-type"))

		return response("201 Created", `{
			"description":"description",
			"displayName": "name",
			"mailNickname": "mail"
		}`)
	})

	client := New(httpClient)

	group, err := client.CreateGroup(context.TODO(), &Group{
		Description:  "description",
		DisplayName:  "name",
		MailNickname: "mail",
	})

	assert.Nil(t, group)
	assert.EqualError(t, err, "azure group 'mail' created, but no ID returned")
}

func Test_GetOrCreateGroupWhenGroupExists(t *testing.T) {
	httpClient := NewTestClient(
		func(req *http.Request) *http.Response {
			assert.Equal(t, "https://graph.microsoft.com/v1.0/groups?%24filter=mailNickname+eq+%27slug%27", req.URL.String())
			assert.Equal(t, http.MethodGet, req.Method)

			return response("200 OK", `{"value": [{"id":"group-id"}]}`)
		},
		func(req *http.Request) *http.Response {
			assert.Fail(t, "Request should not occur")
			return nil
		},
	)

	client := New(httpClient)

	group, created, err := client.GetOrCreateGroup(context.TODO(), "slug", "name", helpers.Strp("description"))

	assert.NoError(t, err)
	assert.Equal(t, "group-id", group.ID)
	assert.False(t, created)
}

func Test_GetOrCreateGroupWhenGroupDoesNotExist(t *testing.T) {
	httpClient := NewTestClient(
		func(req *http.Request) *http.Response {
			assert.Equal(t, "https://graph.microsoft.com/v1.0/groups?%24filter=mailNickname+eq+%27slug%27", req.URL.String())
			assert.Equal(t, http.MethodGet, req.Method)

			return response("404 Not Found", "{}")
		},
		func(req *http.Request) *http.Response {
			assert.Equal(t, "https://graph.microsoft.com/v1.0/groups", req.URL.String())
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, "application/json", req.Header.Get("content-type"))

			return response("201 Created", `{
				"id":"some-id",
				"description":"description",
				"displayName": "name",
				"mailNickname": "mail"
			}`)
		},
	)

	client := New(httpClient)

	group, created, err := client.GetOrCreateGroup(context.TODO(), "slug", "name", helpers.Strp("description"))

	assert.NoError(t, err)
	assert.Equal(t, "some-id", group.ID)
	assert.Equal(t, "description", group.Description)
	assert.Equal(t, "name", group.DisplayName)
	assert.Equal(t, "mail", group.MailNickname)
	assert.True(t, created)
}

func Test_ListGroupMembers(t *testing.T) {
	httpClient := NewTestClient(
		func(req *http.Request) *http.Response {
			assert.Equal(t, "https://graph.microsoft.com/v1.0/groups/group-id/members", req.URL.String())
			assert.Equal(t, http.MethodGet, req.Method)

			return response("200 OK", `{
				"value": [
					{"id": "user-id-1"},
					{"id": "user-id-2"}
				]
			}`)
		},
	)

	client := New(httpClient)

	members, err := client.ListGroupMembers(context.TODO(), &Group{
		ID: "group-id",
	})

	assert.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Equal(t, "user-id-1", members[0].ID)
	assert.Equal(t, "user-id-2", members[1].ID)
}

func Test_ListGroupMembersWhenGroupDoesNotExist(t *testing.T) {
	httpClient := NewTestClient(
		func(req *http.Request) *http.Response {
			assert.Equal(t, "https://graph.microsoft.com/v1.0/groups/group-id/members", req.URL.String())
			assert.Equal(t, http.MethodGet, req.Method)

			return response("404 Not Found", `{"error":{"message":"some error"}}`)
		},
	)

	client := New(httpClient)

	members, err := client.ListGroupMembers(context.TODO(), &Group{
		ID:           "group-id",
		MailNickname: "mail",
	})

	assert.EqualError(t, err, `list group members 'mail': 404 Not Found: {"error":{"message":"some error"}}`)
	assert.Len(t, members, 0)
}

func Test_ListGroupMembersWithInvalidResponse(t *testing.T) {
	httpClient := NewTestClient(
		func(req *http.Request) *http.Response {
			assert.Equal(t, "https://graph.microsoft.com/v1.0/groups/group-id/members", req.URL.String())
			assert.Equal(t, http.MethodGet, req.Method)

			return response("200 OK", "some response")
		},
	)

	client := New(httpClient)

	members, err := client.ListGroupMembers(context.TODO(), &Group{
		ID: "group-id",
	})

	assert.EqualError(t, err, "invalid character 's' looking for beginning of value")
	assert.Nil(t, members)
}

func Test_AddMemberToGroup(t *testing.T) {
	httpClient := NewTestClient(
		func(req *http.Request) *http.Response {
			assert.Equal(t, "https://graph.microsoft.com/v1.0/groups/group-id/members/$ref", req.URL.String())
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, "application/json", req.Header.Get("content-type"))
			body, _ := io.ReadAll(req.Body)
			assert.Equal(t, `{"@odata.id":"https://graph.microsoft.com/v1.0/directoryObjects/user-id"}`, string(body))

			return response("204 No Content", "")
		},
	)

	client := New(httpClient)

	err := client.AddMemberToGroup(context.TODO(), &Group{
		ID: "group-id",
	}, &Member{
		ID:   "user-id",
		Mail: "mail@example.com",
	})

	assert.NoError(t, err)
}

func Test_AddMemberToGroupWithInvalidResponse(t *testing.T) {
	httpClient := NewTestClient(
		func(req *http.Request) *http.Response {
			assert.Equal(t, "https://graph.microsoft.com/v1.0/groups/group-id/members/$ref", req.URL.String())
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, "application/json", req.Header.Get("content-type"))

			return response("200 OK", "some response body")
		},
	)

	client := New(httpClient)

	err := client.AddMemberToGroup(context.TODO(), &Group{
		ID:           "group-id",
		MailNickname: "group",
	}, &Member{
		ID:   "user-id",
		Mail: "mail@example.com",
	})

	assert.EqualError(t, err, "add member 'mail@example.com' to azure group 'group': 200 OK: some response body")
}

func Test_RemoveMemberFromGroup(t *testing.T) {
	httpClient := NewTestClient(
		func(req *http.Request) *http.Response {
			assert.Equal(t, "https://graph.microsoft.com/v1.0/groups/group-id/members/user-id/$ref", req.URL.String())
			assert.Equal(t, http.MethodDelete, req.Method)

			return response("204 No Content", "")
		},
	)

	client := New(httpClient)

	err := client.RemoveMemberFromGroup(context.TODO(), &Group{
		ID: "group-id",
	}, &Member{
		ID: "user-id",
	})

	assert.NoError(t, err)
}

func Test_RemoveMemberFromGroupWithInvalidResponse(t *testing.T) {
	httpClient := NewTestClient(
		func(req *http.Request) *http.Response {
			assert.Equal(t, "https://graph.microsoft.com/v1.0/groups/group-id/members/user-id/$ref", req.URL.String())
			assert.Equal(t, http.MethodDelete, req.Method)

			return response("200 OK", "some response body")
		},
	)

	client := New(httpClient)

	err := client.RemoveMemberFromGroup(context.TODO(), &Group{
		ID:           "group-id",
		MailNickname: "mail@example.com",
	}, &Member{
		ID:   "user-id",
		Mail: "mail",
	})

	assert.EqualError(t, err, "remove member 'mail' from azure group 'mail@example.com': 200 OK: some response body")
}

func response(status string, body string) *http.Response {
	parts := strings.Fields(status)
	statusCode, _ := strconv.Atoi(parts[0])

	return &http.Response{
		Status:     status,
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}
