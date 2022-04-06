package azure_group_reconciler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
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

func (s *azureReconciler) Reconcile(ctx context.Context, in reconcilers.Input) error {
	grp, err := s.getOrCreateGroup(ctx, in)
	if err != nil {
		return err
	}

	// FIXME: support changing metadata (group name, description)

	err = s.connectUsers(ctx, grp, in)
	if err != nil {
		return s.logger.Errorf(in, OpAddMembers, "add members to group: %s", err)
	}

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

// https://docs.microsoft.com/en-us/graph/api/group-list-members?view=graph-rest-1.0&tabs=http
func (s *azureReconciler) listGroupMembers(ctx context.Context, grp *Group) ([]*Member, error) {
	client := s.oauth.Client(ctx)
	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/groups/%s/members", grp.ID)

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
	members := &MemberResponse{}
	err = dec.Decode(members)

	if err != nil {
		return nil, err
	}

	return members.Value, nil
}

func (s *azureReconciler) getOrCreateGroup(ctx context.Context, in reconcilers.Input) (*Group, error) {
	slug := reconcilers.TeamNamePrefix + *in.Team.Slug
	grp, err := s.getGroup(ctx, slug)
	if err == nil {
		return grp, err
	}

	return s.createGroup(ctx, in)
}

func (s *azureReconciler) addMemberToGroup(ctx context.Context, grp *Group, member *Member) error {
	client := s.oauth.Client(ctx)

	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/groups/%s/members/$ref", grp.ID)

	request := &AddMemberRequest{
		ODataID: member.ODataID(),
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")

	resp, err := client.Do(req)
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

func (s *azureReconciler) removeMemberFromGroup(ctx context.Context, grp *Group, member *Member) error {
	client := s.oauth.Client(ctx)

	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/groups/%s/members/%s/$ref", grp.ID, member.ID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", u, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
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

func (s *azureReconciler) connectUsers(ctx context.Context, grp *Group, in reconcilers.Input) error {
	members, err := s.listGroupMembers(ctx, grp)
	if err != nil {
		return s.logger.Errorf(in, OpAddMembers, "list existing members in Azure group: %s", err)
	}

	deleteMembers := extraMembers(members, in.Team.Users)
	createUsers := missingUsers(members, in.Team.Users)

	for _, member := range deleteMembers {
		// FIXME: connect audit log with database user, if exists
		err = s.removeMemberFromGroup(ctx, grp, member)
		if err != nil {
			return s.logger.UserErrorf(in, OpDeleteMember, nil, "delete member '%s' from Azure group '%s': %s", member.Mail, grp.MailNickname, err)
		}
		s.logger.UserLogf(in, OpDeleteMember, nil, "deleted member '%s' from Azure group '%s'", member.Mail, grp.MailNickname)
	}

	s.logger.Logf(in, OpDeleteMembers, "all unmanaged members successfully deleted from Azure group '%s'", grp.MailNickname)

	for _, user := range createUsers {
		if user.Email == nil {
			continue
		}
		member, err := s.getUser(ctx, *user.Email)
		if err != nil {
			return s.logger.UserErrorf(in, OpAddMember, user, "add member '%s' to Azure group '%s': %s", *user.Email, grp.MailNickname, err)
		}
		err = s.addMemberToGroup(ctx, grp, member)
		if err != nil {
			return s.logger.UserErrorf(in, OpAddMember, user, "add member '%s' to Azure group '%s': %s", member.Mail, grp.MailNickname, err)
		}
		s.logger.UserLogf(in, OpAddMember, user, "added member '%s' to Azure group '%s'", member.Mail, grp.MailNickname)
	}

	s.logger.Logf(in, OpAddMembers, "all members successfully added to Azure group '%s'", grp.MailNickname)

	return nil
}

func (s *azureReconciler) getUser(ctx context.Context, email string) (*Member, error) {
	client := s.oauth.Client(ctx)
	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s", email)

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
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

// Given a list of Azure group members and a list of users,
// return users not present in members list.
func missingUsers(members []*Member, users []*dbmodels.User) []*dbmodels.User {
	userMap := make(map[string]*dbmodels.User)
	for _, user := range users {
		if user.Email == nil {
			continue
		}
		userMap[*user.Email] = user
	}
	for _, member := range members {
		delete(userMap, strings.ToLower(member.Mail))
	}
	users = make([]*dbmodels.User, 0, len(userMap))
	for _, user := range userMap {
		users = append(users, user)
	}
	return users
}

// Given a list of Azure group members and a list of users,
// return members not present in user list.
func extraMembers(members []*Member, users []*dbmodels.User) []*Member {
	memberMap := make(map[string]*Member)
	for _, member := range members {
		memberMap[strings.ToLower(member.Mail)] = member
	}
	for _, user := range users {
		if user.Email == nil {
			continue
		}
		delete(memberMap, *user.Email)
	}
	members = make([]*Member, 0, len(memberMap))
	for _, member := range memberMap {
		members = append(members, member)
	}
	return members
}
