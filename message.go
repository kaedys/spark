package spark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

const MessagesURL = "https://api.ciscospark.com/v1/messages"

type Message struct {
	ID          string    `json:"id"`
	RoomID      string    `json:"roomId"`
	RoomType    string    `json:"roomType"`
	PersonID    string    `json:"personId"`
	PersonEmail string    `json:"personEmail"`
	Text        string    `json:"text"`
	Markdown    string    `json:"markdown"`
	Files       []string  `json:"files"`
	HTML        string    `json:"html"`
	Created     time.Time `json:"created"`
}

type MessageList struct {
	Items []*Message
}

// NOTE: One and *only* one of RoomID, ToPersonID, or ToPersonEmail must be set for calls to CreateMessage.
type NewMessage struct {
	RoomID        string   `json:"roomId,omitempty"`
	ToPersonID    string   `json:"toPersonId,omitempty"`
	ToPersonEmail string   `json:"toPersonEmail,omitempty"`
	Text          string   `json:"text,omitempty"`
	Markdown      string   `json:"markdown,omitempty"`
	Files         []string `json:"files,omitempty"`
}

// https://developer.webex.com/endpoint-messages-messageId-get.html
func (c *client) GetMessage(messageID string) (*Message, error) {
	if messageID == "" {
		return nil, fmt.Errorf("no message ID specified")
	}

	resp, err := c.getRequest(fmt.Sprintf("%s/%s", MessagesURL, messageID), nil)
	if err != nil {
		return nil, err
	}

	var m Message
	err = json.Unmarshal(resp, &m)
	return &m, err
}

// https://developer.webex.com/endpoint-messages-post.html
func (c *client) CreateMessage(m *NewMessage) (*Message, error) {
	if m == nil {
		return nil, fmt.Errorf("nil message")
	}
	if m.RoomID == "" && m.ToPersonEmail == "" && m.ToPersonID == "" {
		return nil, fmt.Errorf("message requires a room ID, person ID, or email to send to")
	}

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(m); err != nil {
		return nil, err
	}
	resp, err := c.postRequest(MessagesURL, b)
	if err != nil {
		return nil, err
	}

	var rm Message
	err = json.Unmarshal(resp, &rm)
	return &rm, err
}

// https://developer.webex.com/endpoint-messages-messageId-delete.html
func (c *client) DeleteMessage(messageID string) error {
	if messageID == "" {
		return fmt.Errorf("no message ID specified")
	}

	_, err := c.deleteRequest(fmt.Sprintf("%s/%s", MessagesURL, messageID))
	return err
}

// https://developer.ciscospark.com/endpoint-messages-get.html
func (c *client) ListMessages(max int, roomID string, params *MessageListParams) ([]*Message, error) {
	if roomID == "" {
		return nil, fmt.Errorf("no room ID specified")
	}

	resp, reqErr := c.getRequestWithPaging(MessagesURL, params.values(roomID), max)
	if reqErr != nil && len(resp) == 0 { // if we got an error *and* results, parse them and return them
		return nil, reqErr
	}

	var messages []*Message
	for _, r := range resp {
		var ml MessageList
		if jsonErr := json.Unmarshal(r, &ml); reqErr != nil {
			return messages, fmt.Errorf("%v && %v", reqErr, jsonErr)
		}
		messages = append(messages, ml.Items...)
	}
	return messages, reqErr
}

type MessageListParams struct {
	MentionedPeople string
	Before          time.Time
	BeforeMessageID string
}

func (m *MessageListParams) values(roomID string) url.Values {
	uv := make(url.Values)
	uv.Add("roomId", roomID)

	if m == nil {
		return uv
	}

	if m.MentionedPeople != "" {
		uv.Add("mentionedPeople", m.MentionedPeople)
	}
	if m.Before != (time.Time{}) { // zero value
		uv.Add("before", m.Before.Format(time.RFC3339))
	}
	if m.BeforeMessageID != "" {
		uv.Add("beforeMessage", m.BeforeMessageID)
	}

	return uv
}
