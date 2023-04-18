package dependencytrack

/*
type mock struct {
	teams       map[string]*Team
	permissions map[string][]*Permission
	users       []*User
}

func newMock() *mock {
	return &mock{
		teams:       make(map[string]*Team),
		permissions: make(map[string][]*Permission),
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
	m.teams[team.Uuid] = team
	m.permissions[teamName] = per
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
	var user *User
	for _, u := range m.users {
		if u.Username == username {
			user = u
		}
	}
	m.teams[uuid].OidcUsers = append(m.teams[uuid].OidcUsers, *user)
	return nil
}

func (m *mock) DeleteTeam(_ context.Context, uuid string) error {
	m.teams[uuid] = nil
	return nil
}

func (m *mock) DeleteUserMembership(_ context.Context, uuid string, username string) error {
	users := m.teams[uuid].OidcUsers
	for i, u := range users {
		if u.Username == username {
			users = append(users[:i], users[i+1:]...)
		}
	}
	return nil
}

func (m *mock) reset() {
	m.teams = make(map[string]*Team, 0)
	m.permissions = map[string][]*Permission{}
	m.users = make([]*User, 0)
}

func (m *mock) membersInTeam(teamName string) []string {
	users := make([]string, 0)
	for _, t := range m.teams {
		if t.Name == teamName {
			for _, u := range t.OidcUsers {
				users = append(users, u.Username)
			}
		}
	}
	return users
}

func (m *mock) usernames() []string {
	users := make([]string, 0)
	for _, u := range m.users {
		users = append(users, u.Username)
	}
	return users
}

func (m *mock) teamExist(uuid string) bool {
	for _, t := range m.teams {
		if t.Uuid == uuid {
			return true
		}
	}
	return false
}

func (m *mock) hasPermissions(team string, permissions ...Permission) bool {
	for _, p := range m.permissions[team] {
		for _, p2 := range permissions {
			if *p != p2 {
				return false
			}
		}
	}
	return true
}
*/
