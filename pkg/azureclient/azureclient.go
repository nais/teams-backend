package azureclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/reconcilers"
	"io"
	"net/http"
)

type client struct {
	client *http.Client
}

type Client interface {
	AddMemberToGroup(ctx context.Context, grp *Group, member *Member) error
	CreateGroup(ctx context.Context, grp *Group) (*Group, error)
	GetGroupById(ctx context.Context, id uuid.UUID) (*Group, error)
	GetOrCreateGroup(ctx context.Context, state reconcilers.AzureState, slug, name string, description *string) (*Group, bool, error)
	GetUser(ctx context.Context, email string) (*Member, error)
	ListGroupMembers(ctx context.Context, grp *Group) ([]*Member, error)
	ListGroupOwners(ctx context.Context, grp *Group) ([]*Owner, error)
	RemoveMemberFromGroup(ctx context.Context, grp *Group, member *Member) error
}

func New(c *http.Client) Client {
	return &client{
		client: c,
	}
}

func (s *client) GetUser(ctx context.Context, email string) (*Member, error) {
	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s", email)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		text, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s: %s", resp.Status, string(text))
	}

	dec := json.NewDecoder(resp.Body)
	user := &Member{}
	err = dec.Decode(user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *client) GetGroupById(ctx context.Context, id uuid.UUID) (*Group, error) {
	u := "https://graph.microsoft.com/v1.0/groups/" + id.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("azure group with ID '%s' does not exist", id.String())
	}

	dec := json.NewDecoder(resp.Body)
	grp := &Group{}
	err = dec.Decode(grp)

	if err != nil {
		return nil, err
	}

	return grp, nil
}

func (s *client) CreateGroup(ctx context.Context, grp *Group) (*Group, error) {
	u := "https://graph.microsoft.com/v1.0/groups"

	payload, err := json.Marshal(grp)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		text, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create azure group '%s': %s: %s", grp.MailNickname, resp.Status, string(text))
	}

	dec := json.NewDecoder(resp.Body)
	grp = &Group{}
	err = dec.Decode(grp)

	if err != nil {
		return nil, err
	}

	if len(grp.ID) == 0 {
		return nil, fmt.Errorf("azure group '%s' created, but no ID returned", grp.MailNickname)
	}

	return grp, nil
}

// GetOrCreateGroup Get or create a group fom the Graph API. The second return value informs if the group was
// created or not.
func (s *client) GetOrCreateGroup(ctx context.Context, state reconcilers.AzureState, mailNickname, name string, description *string) (*Group, bool, error) {
	if state.GroupID != nil {
		grp, err := s.GetGroupById(ctx, *state.GroupID)
		if err == nil {
			return grp, false, err
		}
	}

	createdGroup, err := s.CreateGroup(ctx, &Group{
		Description:     helpers.DerefString(description),
		DisplayName:     name,
		GroupTypes:      nil,
		MailEnabled:     false,
		MailNickname:    mailNickname,
		SecurityEnabled: true,
	})
	if err != nil {
		return nil, false, err
	}
	return createdGroup, true, nil
}

// https://docs.microsoft.com/en-us/graph/api/group-list-owners?view=graph-rest-1.0&tabs=http
func (s *client) ListGroupOwners(ctx context.Context, grp *Group) ([]*Owner, error) {
	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/groups/%s/owners", grp.ID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		text, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list group owners '%s': %s: %s", grp.MailNickname, resp.Status, string(text))
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	owners := &OwnerResponse{}
	err = dec.Decode(owners)

	if err != nil {
		return nil, err
	}

	return owners.Value, nil
}

// https://docs.microsoft.com/en-us/graph/api/group-list-members?view=graph-rest-1.0&tabs=http
func (s *client) ListGroupMembers(ctx context.Context, grp *Group) ([]*Member, error) {
	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/groups/%s/members", grp.ID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		text, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list group members '%s': %s: %s", grp.MailNickname, resp.Status, string(text))
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	members := &MemberResponse{}
	err = dec.Decode(members)

	if err != nil {
		return nil, err
	}

	return members.Value, nil
}

func (s *client) AddMemberToGroup(ctx context.Context, grp *Group, member *Member) error {
	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/groups/%s/members/$ref", grp.ID)

	request := &AddMemberRequest{
		ODataID: member.ODataID(),
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		text, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add member '%s' to azure group '%s': %s: %s", member.Mail, grp.MailNickname, resp.Status, string(text))
	}

	return nil
}

func (s *client) RemoveMemberFromGroup(ctx context.Context, grp *Group, member *Member) error {
	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/groups/%s/members/%s/$ref", grp.ID, member.ID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		text, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remove member '%s' from azure group '%s': %s: %s", member.Mail, grp.MailNickname, resp.Status, string(text))
	}

	return nil
}
