package spark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

const RoomsURL = "https://api.ciscospark.com/v1/rooms"

type Room struct {
	ID           string    `json:"id,omitempty"`
	Title        string    `json:"title,omitempty"`
	Type         string    `json:"type,omitempty"`
	IsLocked     bool      `json:"isLocked,omitempty"`
	SIPAddress   string    `json:"sipAddress,omitempty"`
	TeamID       string    `json:"teamId,omitempty"`
	LastActivity time.Time `json:"lastActivity,omitempty"`
	CreatorID    string    `json:"creatorId,omitempty"`
	Created      time.Time `json:"created,omitempty"`
}

type RoomList struct {
	Items []*Room
}

// https://developer.webex.com/endpoint-rooms-roomId-get.html
func (c *client) GetRoom(roomId string) (*Room, error) {
	if roomId == "" {
		return nil, fmt.Errorf("no room ID specified")
	}
	resp, err := c.getRequest(fmt.Sprintf("%s/%s", RoomsURL, roomId), nil)
	if err != nil {
		return nil, err
	}

	var room Room
	err = json.Unmarshal(resp, &room)
	return &room, err
}

// GetRoomByName is a helper method that wraps GetRoom.  It will query for all rooms that the user is a member of, then
// return the first one that matches the provided name.  If no such room exists, an error will be returned instead.
func (c *client) GetRoomByName(roomName string) (*Room, error) {
	if roomName == "" {
		return nil, fmt.Errorf("no room name specified")
	}

	allRooms, err := c.ListRooms(0, nil)
	if err != nil {
		return nil, err
	}

	for _, r := range allRooms {
		if r.Title == roomName {
			return r, nil
		}
	}

	return nil, fmt.Errorf("no room with name %q was found", roomName)
}

// https://developer.webex.com/endpoint-rooms-post.html
func (c *client) CreateRoom(name, teamID string) (*Room, error) {
	if name == "" {
		return nil, fmt.Errorf("no room name specified")
	}
	// weirdly, a team ID isn't required

	r := Room{
		Title:  name,
		TeamID: teamID,
	}

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(r); err != nil {
		return nil, err
	}
	resp, err := c.postRequest(RoomsURL, b)
	if err != nil {
		return nil, err
	}

	var rr Room
	err = json.Unmarshal(resp, &rr)
	return &rr, err
}

// https://developer.webex.com/endpoint-rooms-roomId-put.html
func (c *client) UpdateRoomName(roomID, newName string) (*Room, error) {
	if roomID == "" {
		return nil, fmt.Errorf("no room ID specified")
	}
	if newName == "" {
		return nil, fmt.Errorf("no room name specified")
	}

	r := Room{Title: newName}

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(r); err != nil {
		return nil, err
	}
	resp, err := c.putRequest(fmt.Sprintf("%s/%s", RoomsURL, roomID), b)
	if err != nil {
		return nil, err
	}

	var rr Room
	err = json.Unmarshal(resp, &rr)
	return &rr, err
}

// https://developer.webex.com/endpoint-rooms-roomId-delete.html
func (c *client) DeleteRoom(roomID string) error {
	if roomID == "" {
		return fmt.Errorf("no room ID specified")
	}

	_, err := c.deleteRequest(fmt.Sprintf("%s/%s", RoomsURL, roomID))
	return err
}

// https://developer.webex.com/endpoint-rooms-get.html
func (c *client) ListRooms(max int, params *RoomListParams) ([]*Room, error) {
	resp, reqErr := c.getRequestWithPaging(RoomsURL, params.values(), max)
	if reqErr != nil && len(resp) == 0 { // if we got an error *and* results, parse them and return them
		return nil, reqErr
	}

	var rooms []*Room
	for _, r := range resp {
		var rl RoomList
		if jsonErr := json.Unmarshal(r, &rl); jsonErr != nil {
			return rooms, fmt.Errorf("%v && %v", reqErr, jsonErr)
		}
		rooms = append(rooms, rl.Items...)
	}
	return rooms, nil
}

type RoomListParams struct {
	TeamID string
	Type   string
	SortBy string
}

func (r *RoomListParams) values() url.Values {
	uv := make(url.Values)
	if r == nil {
		return uv
	}

	if r.TeamID != "" {
		uv.Add("teamId", r.TeamID)
	}
	if r.Type != "" {
		uv.Add("type", r.Type)
	}
	if r.SortBy != "" {
		uv.Add("sortBy", r.SortBy)
	}

	return uv
}
