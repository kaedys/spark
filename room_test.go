package spark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Room (Live)", func() {
	var c Client

	BeforeEach(func() {
		httpCli = httpClient(new(http.Client))

		if token == "" {
			Skip("token not set, skipping live tests")
		}
		c = New(token)
	})

	Describe("GetRoom", func() {
		var room *Room
		BeforeEach(func() {
			r, err := c.ListRooms(5, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(r).ToNot(BeEmpty())
			room = r[0]
		})

		It("gets a room by ID", func() {
			Expect(c.GetRoom(room.ID)).To(Equal(room))
		})

		It("gets a room by name", func() {
			Expect(c.GetRoomByName(room.Title)).To(Equal(room))
		})
	})

	Describe("ListRooms", func() {
		It("gets a list of rooms", func() {
			Expect(c.ListRooms(5, nil)).ToNot(BeEmpty())
		})
	})
})

var _ = Describe("Room (Mock)", func() {
	var c Client
	var mockCli *mockHTTPClient

	var rooms RoomList

	BeforeEach(func() {
		c = New("mock")
		mockCli = new(mockHTTPClient)
		httpCli = mockCli // set client global to a mock

		rooms = RoomList{
			Items: []*Room{
				{
					ID:     "1",
					Title:  "room 1",
					TeamID: "team 1",
				},
				{
					ID:     "2",
					Title:  "room 2",
					TeamID: "team 2",
				},
				{
					ID:     "2",
					Title:  "room 2",
					TeamID: "team 2",
				},
			},
		}
	})

	Describe("GetRoom", func() {
		It("gets a room by ID", func() {
			roomID := rooms.Items[0].ID

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/%s", RoomsURL, roomID)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(rooms.Items[0])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.GetRoom(roomID)).To(Equal(rooms.Items[0]))
		})

		It("fails if no room ID is specified", func() {
			p, err := c.GetRoom("")
			Expect(err).To(MatchError("no room ID specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.GetRoom("1")
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("GetRoomByName", func() {
		It("gets a room by name", func() {
			roomName := rooms.Items[0].Title

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(RoomsURL))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(rooms)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.GetRoomByName(roomName)).To(Equal(rooms.Items[0]))
		})

		It("fails if the room can't be found", func() {
			roomName := "not a room"

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(RoomsURL))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(rooms)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			p, err := c.GetRoomByName(roomName)
			Expect(err).To(MatchError(fmt.Sprintf("no room with name %q was found", roomName)))
			Expect(p).To(BeNil())
		})

		It("fails if no room name is specified", func() {
			p, err := c.GetRoomByName("")
			Expect(err).To(MatchError("no room name specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.GetRoomByName("1")
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("ListRooms", func() {
		It("gets a list of rooms", func() {
			max := len(rooms.Items)

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(RoomsURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(rooms)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListRooms(max, nil)).To(ConsistOf(rooms.Items))
		})

		It("gets a list of rooms with a maximum", func() {
			max := len(rooms.Items) - 1

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(RoomsURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				rooms.Items = rooms.Items[:max]

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(rooms)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListRooms(max, nil)).To(ConsistOf(rooms.Items))
		})

		It("sets max parameter to the client max if max arg = 0", func() {
			max := 0
			cmax := 25
			c = c.SetMaxPerPage(cmax)

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(RoomsURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(rooms)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListRooms(max, nil)).To(ConsistOf(rooms.Items))
		})

		It("pages if max > client max", func() {
			max := len(rooms.Items)
			cmax := 1
			c = c.SetMaxPerPage(cmax)

			calls := 0
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(RoomsURL))
				if calls == 0 {
					Expect(req.URL.Query().Get("after")).To(BeEmpty())
				} else {
					Expect(req.URL.Query().Get("after")).To(Equal(rooms.Items[calls-1].ID))
				}
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				p := RoomList{
					Items: rooms.Items[calls : calls+1],
				}

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(p)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}

				if calls < max {
					r.Header = map[string][]string{
						"Link": {fmt.Sprintf("<%s?max=%d&after=%s>; rel=\"next\"", RoomsURL, cmax, rooms.Items[calls].ID)},
					}
				}

				calls++

				return r, nil
			}

			Expect(c.ListRooms(max, nil)).To(ConsistOf(rooms.Items))
			Expect(calls).To(BeEquivalentTo(3))
		})

		It("pages until it stops getting next links if max = 0", func() {
			max := 0
			cmax := len(rooms.Items)
			c = c.SetMaxPerPage(cmax)

			calls := 0
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(RoomsURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(rooms)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				calls++

				if calls < 10 {
					r.Header = map[string][]string{
						"Link": {fmt.Sprintf("<%s>; rel=\"next\"", RoomsURL)},
					}
				}

				return r, nil
			}

			Expect(c.ListRooms(max, nil)).To(HaveLen(len(rooms.Items) * 10))
			Expect(calls).To(BeEquivalentTo(10))
		})

		It("applies a parameter list", func() {
			max := len(rooms.Items)
			params := RoomListParams{
				TeamID: "test team ID",
				Type:   "test type",
				SortBy: "test sort by",
			}

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(RoomsURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				for k, v := range params.values() {
					Expect(req.URL.Query().Get(k)).To(Equal(v[0]), fmt.Sprintf("MISSING [%s] %+v", k, req.URL.Query()))
				}

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(rooms)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListRooms(max, &params)).To(ConsistOf(rooms.Items))
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.ListRooms(0, nil)
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("CreateRoom", func() {
		It("creates a room", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(RoomsURL))
				Expect(req.Method).To(Equal("POST"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var p Room
				Expect(json.NewDecoder(req.Body).Decode(&p)).To(Succeed())
				Expect(p.Title).To(Equal(rooms.Items[0].Title))
				Expect(p.TeamID).To(Equal(rooms.Items[0].TeamID))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(rooms.Items[1])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.CreateRoom(rooms.Items[0].Title, rooms.Items[0].TeamID)).To(Equal(rooms.Items[1]))
		})

		It("fails if an empty room name is provided", func() {
			p, err := c.CreateRoom("", "")
			Expect(err).To(MatchError("no room name specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.CreateRoom("1", "")
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("UpdateRoomName", func() {
		It("updates a room name", func() {
			newName := "new room name"
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/%s", RoomsURL, rooms.Items[0].ID)))
				Expect(req.Method).To(Equal("PUT"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var p Room
				Expect(json.NewDecoder(req.Body).Decode(&p)).To(Succeed())
				Expect(p.Title).To(Equal(newName))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(rooms.Items[1])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.UpdateRoomName(rooms.Items[0].ID, newName)).To(Equal(rooms.Items[1]))
		})

		It("fails if an empty room ID is provided", func() {
			p, err := c.UpdateRoomName("", "1")
			Expect(err).To(MatchError("no room ID specified"))
			Expect(p).To(BeNil())
		})

		It("fails if an empty room name is provided", func() {
			p, err := c.UpdateRoomName("1", "")
			Expect(err).To(MatchError("no room name specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.UpdateRoomName("1", "2")
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("DeleteRoom", func() {
		It("deletes a room", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/%s", RoomsURL, rooms.Items[0].ID)))
				Expect(req.Method).To(Equal("DELETE"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				r := &http.Response{
					Body:       closer(&bytes.Buffer{}), // empty body
					StatusCode: http.StatusNoContent,    // deletes are weird and return 204 instead of 200
				}
				return r, nil
			}

			Expect(c.DeleteRoom(rooms.Items[0].ID)).To(Succeed())
		})

		It("fails if the room ID is empty", func() {
			Expect(c.DeleteRoom("")).To(MatchError("no room ID specified"))
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			Expect(c.DeleteRoom("1")).To(MatchError(mockErr))
		})
	})
})
