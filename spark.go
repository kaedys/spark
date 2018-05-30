package spark

type Client interface {
	SetMaxPerPage(max int) Client

	GetPerson(personID string) (*Person, error)
	GetMyself() (*Person, error)
	ListPeople(max int, params *PeopleListParams) ([]*Person, error)
	CreatePerson(p *Person) (*Person, error)
	UpdatePerson(p *Person) (*Person, error)
	DeletePerson(ID string) error

	GetRoom(roomId string) (*Room, error)
	GetRoomByName(roomName string) (*Room, error)
	ListRooms(max int, params *RoomListParams) ([]*Room, error)
	CreateRoom(name, teamID string) (*Room, error)
	UpdateRoomName(roomID, newName string) (*Room, error)
	DeleteRoom(roomID string) error

	GetMessage(messageID string) (*Message, error)
	ListMessages(max int, roomID string, params *MessageListParams) ([]*Message, error)
	CreateMessage(m *NewMessage) (*Message, error)
	DeleteMessage(messageID string) error

	GetWebhook(webhookID string) (*Webhook, error)
	ListWebhooks(max int) ([]*Webhook, error)
	CreateWebhook(w *NewWebhook) (*Webhook, error)
	UpdateWebhook(w *Webhook) (*Webhook, error)
	DeleteWebhook(hookID string) error
}

type client struct {
	token   string
	pageMax int
}

func New(token string) Client {
	return &client{
		token:   token,
		pageMax: 50,
	}
}

// Sets the maximum entries per page for paginated queries.  Does not modify the calling client.  Instead, returns
// a *copy* of the calling client with the new max, so it can be daisychained into further calls. Ex:
//
//   cli.SetMaxPerPage(25).ListPeople(50, nil)
//
// To set the value permanently on a new client, daisychain it on to the New() call:
//
//   cli := spark.New(token).SetMaxPerPage(25)
//
func (c *client) SetMaxPerPage(max int) Client {
	return &client{
		token:   c.token,
		pageMax: max,
	}
}
