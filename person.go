package spark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

const PeopleURL = "https://api.ciscospark.com/v1/people"

type Person struct {
	ID            string    `json:"id,omitempty"`
	Emails        []string  `json:"emails,omitempty"`
	DisplayName   string    `json:"displayName,omitempty"`
	NickName      string    `json:"nickName,omitempty"`
	FirstName     string    `json:"firstName,omitempty"`
	LastName      string    `json:"lastName,omitempty"`
	Avatar        string    `json:"avatar,omitempty"`
	OrgId         string    `json:"orgId,omitempty"`
	Roles         []string  `json:"roles,omitempty"`
	Licenses      []string  `json:"licenses,omitempty"`
	Created       time.Time `json:"created,omitempty"`
	Timezone      string    `json:"timezone,omitempty"`
	LastActivity  time.Time `json:"lastActivity,omitempty"`
	Status        string    `json:"status,omitempty"`
	InvitePending bool      `json:"invitePending,omitempty"`
	LoginEnabled  bool      `json:"loginEnabled,omitempty"`
	Type          string    `json:"type,omitempty"`
}

type People struct {
	Items []*Person
}

// https://developer.webex.com/endpoint-people-personId-get.html
func (c *client) GetPerson(personID string) (*Person, error) {
	if personID == "" {
		return nil, fmt.Errorf("no person ID specified")
	}

	resp, err := c.getRequest(fmt.Sprintf("%s/%s", PeopleURL, personID), nil)
	if err != nil {
		return nil, err
	}

	var person Person
	if err := json.Unmarshal(resp, &person); err != nil {
		return nil, err
	}
	return &person, err
}

// https://developer.webex.com/endpoint-people-me-get.html
func (c *client) GetMyself() (*Person, error) {
	return c.GetPerson("me")
}

// https://developer.webex.com/endpoint-people-post.html
func (c *client) CreatePerson(p *Person) (*Person, error) {
	if p == nil {
		return nil, fmt.Errorf("nil person")
	}
	if len(p.Emails) == 0 { // strangely, the only required field
		return nil, fmt.Errorf("no email specified")
	}

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(p); err != nil {
		return nil, err
	}
	resp, err := c.postRequest(PeopleURL, b)
	if err != nil {
		return nil, err
	}

	var rp Person
	err = json.Unmarshal(resp, &rp)
	return &rp, err
}

// https://developer.webex.com/endpoint-people-personId-put.html
func (c *client) UpdatePerson(p *Person) (*Person, error) {
	if p == nil {
		return nil, fmt.Errorf("nil person")
	}
	if p.ID == "" {
		return nil, fmt.Errorf("no person ID specified")
	}
	// weirdly, Emails isn't required, despite the fact that it's required for a *new* person

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(p); err != nil {
		return nil, err
	}
	resp, err := c.putRequest(fmt.Sprintf("%s/%s", PeopleURL, p.ID), b)
	if err != nil {
		return nil, err
	}

	var rp Person
	err = json.Unmarshal(resp, &rp)
	return &rp, err
}

// https://developer.webex.com/endpoint-people-personId-delete.html
func (c *client) DeletePerson(ID string) error {
	if ID == "" {
		return fmt.Errorf("no person ID specified")
	}

	_, err := c.deleteRequest(fmt.Sprintf("%s/%s", PeopleURL, ID))
	return err
}

// https://developer.webex.com/endpoint-people-get.html
func (c *client) ListPeople(max int, params *PeopleListParams) ([]*Person, error) {
	resp, reqErr := c.getRequestWithPaging(PeopleURL, params.values(), max)
	if reqErr != nil && len(resp) == 0 { // if we got an error *and* results, parse them and return them
		return nil, reqErr
	}

	var people []*Person
	for _, r := range resp {
		var pl People
		if jsonErr := json.Unmarshal(r, &pl); jsonErr != nil {
			return people, fmt.Errorf("%v && %v", reqErr, jsonErr)
		}
		people = append(people, pl.Items...)
	}
	return people, nil
}

type PeopleListParams struct {
	Email       string
	DisplayName string
	ID          string
	OrgID       string
}

func (p *PeopleListParams) values() url.Values {
	uv := make(url.Values)
	if p == nil {
		return uv
	}

	if p.Email != "" {
		uv.Add("email", p.Email)
	}
	if p.DisplayName != "" {
		uv.Add("displayName", p.DisplayName)
	}
	if p.ID != "" {
		uv.Add("id", p.ID)
	}
	if p.OrgID != "" {
		uv.Add("orgId", p.OrgID)
	}

	return uv
}
