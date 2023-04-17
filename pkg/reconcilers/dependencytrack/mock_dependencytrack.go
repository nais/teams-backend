package dependencytrack

import (
	"context"
	"github.com/google/uuid"
)

type mock struct {
	teams       []*Team
	Members     map[string][]string
	Permissions map[string][]*Permission
	users       []*User
}

type mockTeam struct {
	*Team
	Members     []*User
	Permissions []*Permission
}

func newMock() *mock {
	return &mock{
		teams:       make([]*Team, 0),
		Members:     make(map[string][]string),
		Permissions: map[string][]*Permission{},
		users:       make([]*User, 0),
	}
}

func (m *mock) CreateTeam(ctx context.Context, teamName string, permissions []Permission) (*Team, error) {
	per := make([]*Permission, 0)
	for _, p := range permissions {
		per = append(per, &p)
	}
	team := &Team{
		Name: teamName,
		Uuid: uuid.New().String(),
	}
	m.teams = append(m.teams, team)
	m.Permissions[teamName] = per
	return team, nil
}

func (m *mock) GetTeams(ctx context.Context) ([]Team, error) {
	t := make([]Team, 0)
	for _, team := range m.teams {
		t = append(t, *team)
	}
	return t, nil
}

func (m *mock) CreateUser(ctx context.Context, email string) error {
	u := &User{
		Email:    email,
		Username: email,
	}

	m.users = append(m.users, u)
	return nil
}

func (m *mock) AddToTeam(ctx context.Context, username string, uuid string) error {
	m.Members[uuid] = append(m.Members[uuid], username)
	return nil
}

func (m *mock) DeleteTeam(ctx context.Context, uuid string) error {
	for _, t := range m.teams {
		if t.Uuid == uuid {
			m.teams = append(m.teams[:0], m.teams[1:]...)
		}
	}
	return nil
}

func (m *mock) DeleteUserMembership(ctx context.Context, uuid string, username string) error {
	// TODO: check that it really works
	for _, t := range m.Members[uuid] {
		if t == username {
			m.Members[uuid] = append(m.Members[uuid][:0], m.Members[uuid][1:]...)
		}
	}
	return nil
}
