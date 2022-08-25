package azureclient

type GroupResponse struct {
	Value []*Group
}

type Group struct {
	ID              string   `json:"id,omitempty"`
	Description     string   `json:"description,omitempty"`
	DisplayName     string   `json:"displayName,omitempty"`
	GroupTypes      []string `json:"groupTypes,omitempty"`
	MailEnabled     bool     `json:"mailEnabled"`
	MailNickname    string   `json:"mailNickname,omitempty"`
	SecurityEnabled bool     `json:"securityEnabled"`
}

type MemberResponse struct {
	Value []*Member
}

type Member struct {
	ID        string `json:"id,omitempty"`
	GivenName string `json:"givenNAme,omitempty"`
	Surname   string `json:"surname,omitempty"`
	Mail      string `json:"mail,omitempty"`
}

type AddMemberRequest struct {
	ODataID string `json:"@odata.id"`
}

func (m Member) ODataID() string {
	return "https://graph.microsoft.com/v1.0/directoryObjects/" + m.ID
}

func (m Member) Name() string {
	return m.GivenName + " " + m.Surname
}
