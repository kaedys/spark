package spark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"strings"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Message (Mock)", func() {
	var c Client
	var mockCli *mockHTTPClient

	var messages MessageList

	BeforeEach(func() {
		c = New("mock")
		mockCli = new(mockHTTPClient)
		httpCli = mockCli // set client global to a mock

		messages = MessageList{
			Items: []*Message{
				{
					ID:          "1",
					RoomID:      "room ID 1",
					RoomType:    "room type 1",
					PersonID:    "person ID 1",
					PersonEmail: "person email 1",
					Markdown:    "markdown 1",
				},
				{
					ID:          "2",
					RoomID:      "room ID 2",
					RoomType:    "room type 2",
					PersonID:    "person ID 2",
					PersonEmail: "person email 2",
					Markdown:    "markdown 2",
				},
				{
					ID:          "3",
					RoomID:      "room ID 3",
					RoomType:    "room type 3",
					PersonID:    "person ID 3",
					PersonEmail: "person email 3",
					Markdown:    "markdown 3",
				},
			},
		}
	})

	Describe("GetMessage", func() {
		It("gets a message by ID", func() {
			messageID := messages.Items[0].ID

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/%s", MessagesURL, messageID)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(messages.Items[0])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.GetMessage(messageID)).To(Equal(messages.Items[0]))
		})

		It("fails if no message ID is specified", func() {
			p, err := c.GetMessage("")
			Expect(err).To(MatchError("no message ID specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.GetMessage("1")
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("ListMessages", func() {
		It("gets a list of messages", func() {
			max := len(messages.Items)
			roomID := "123"

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(MessagesURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				Expect(req.URL.Query().Get("roomId")).To(Equal(roomID))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(messages)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListMessages(max, roomID, nil)).To(ConsistOf(messages.Items))
		})

		It("gets a list of messages with a maximum", func() {
			max := len(messages.Items) - 1
			roomID := "123"

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(MessagesURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				Expect(req.URL.Query().Get("roomId")).To(Equal(roomID))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				messages.Items = messages.Items[:max]

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(messages)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListMessages(max, roomID, nil)).To(ConsistOf(messages.Items))
		})

		It("sets max parameter to the client max if max arg = 0", func() {
			max := 0
			cmax := 25
			c = c.SetMaxPerPage(cmax)
			roomID := "123"

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(MessagesURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.URL.Query().Get("roomId")).To(Equal(roomID))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(messages)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListMessages(max, roomID, nil)).To(ConsistOf(messages.Items))
		})

		It("pages if max > client max", func() {
			max := len(messages.Items)
			cmax := 1
			c = c.SetMaxPerPage(cmax)
			roomID := "123"

			calls := 0
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(MessagesURL))
				if calls == 0 {
					Expect(req.URL.Query().Get("after")).To(BeEmpty())
				} else {
					Expect(req.URL.Query().Get("after")).To(Equal(messages.Items[calls-1].ID))
				}
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.URL.Query().Get("roomId")).To(Equal(roomID))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				p := MessageList{
					Items: messages.Items[calls : calls+1],
				}

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(p)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}

				if calls < max {
					r.Header = map[string][]string{
						"Link": {fmt.Sprintf("<%s?max=%d&after=%s>; rel=\"next\"", MessagesURL, cmax, messages.Items[calls].ID)},
					}
				}

				calls++

				return r, nil
			}

			Expect(c.ListMessages(max, roomID, nil)).To(ConsistOf(messages.Items))
			Expect(calls).To(BeEquivalentTo(3))
		})

		It("pages until it stops getting next links if max = 0", func() {
			max := 0
			cmax := len(messages.Items)
			c = c.SetMaxPerPage(cmax)
			roomID := "123"

			calls := 0
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(MessagesURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.URL.Query().Get("roomId")).To(Equal(roomID))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(messages)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				calls++

				if calls < 10 {
					r.Header = map[string][]string{
						"Link": {fmt.Sprintf("<%s>; rel=\"next\"", MessagesURL)},
					}
				}

				return r, nil
			}

			Expect(c.ListMessages(max, roomID, nil)).To(HaveLen(len(messages.Items) * 10))
			Expect(calls).To(BeEquivalentTo(10))
		})

		It("applies a parameter list", func() {
			max := len(messages.Items)
			params := MessageListParams{
				MentionedPeople: "mentioned",
				Before:          time.Now(),
				BeforeMessageID: "befoire",
			}
			roomID := "123"

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(MessagesURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				Expect(req.URL.Query().Get("roomId")).To(Equal(roomID))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				for k, v := range params.values(roomID) {
					Expect(req.URL.Query().Get(k)).To(Equal(v[0]), fmt.Sprintf("MISSING [%s] %+v", k, req.URL.Query()))
				}

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(messages)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListMessages(max, roomID, &params)).To(ConsistOf(messages.Items))
		})

		It("fails if an empty room ID is provided", func() {
			p, err := c.ListMessages(0, "", nil)
			Expect(err).To(MatchError("no room ID specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.ListMessages(0, "123", nil)
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("CreateMessage", func() {
		var n NewMessage

		BeforeEach(func() {
			// Wash it through the json package, because honestly that's the easiest way to copy a struct to
			// a struct with a subset of the same fields
			var b bytes.Buffer
			Expect(json.NewEncoder(&b).Encode(messages.Items[0])).To(Succeed())
			Expect(json.NewDecoder(&b).Decode(&n)).To(Succeed())
			n.ToPersonID = messages.Items[0].PersonID
			n.ToPersonEmail = messages.Items[0].PersonEmail

			Expect(n.RoomID).To(Equal(messages.Items[0].RoomID))
			Expect(n.ToPersonID).To(Equal(messages.Items[0].PersonID))
			Expect(n.ToPersonEmail).To(Equal(messages.Items[0].PersonEmail))
			Expect(n.Text).To(Equal(messages.Items[0].Text))
			Expect(n.Markdown).To(Equal(messages.Items[0].Markdown))
			Expect(n.Files).To(Equal(messages.Items[0].Files))
		})

		It("creates a message", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(MessagesURL))
				Expect(req.Method).To(Equal("POST"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var p NewMessage
				Expect(json.NewDecoder(req.Body).Decode(&p)).To(Succeed())
				Expect(p).To(Equal(n))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(messages.Items[1])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.CreateMessage(&n)).To(Equal(messages.Items[1]))
		})

		It("allows an empty room ID", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(MessagesURL))
				Expect(req.Method).To(Equal("POST"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var p NewMessage
				Expect(json.NewDecoder(req.Body).Decode(&p)).To(Succeed())
				Expect(p).To(Equal(n))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(messages.Items[1])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}
			n.RoomID = ""

			Expect(c.CreateMessage(&n)).To(Equal(messages.Items[1]))
		})

		It("allows an empty person email", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(MessagesURL))
				Expect(req.Method).To(Equal("POST"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var p NewMessage
				Expect(json.NewDecoder(req.Body).Decode(&p)).To(Succeed())
				Expect(p).To(Equal(n))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(messages.Items[1])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}
			n.ToPersonEmail = ""

			Expect(c.CreateMessage(&n)).To(Equal(messages.Items[1]))
		})

		It("allows an empty person ID", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(MessagesURL))
				Expect(req.Method).To(Equal("POST"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var p NewMessage
				Expect(json.NewDecoder(req.Body).Decode(&p)).To(Succeed())
				Expect(p).To(Equal(n))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(messages.Items[1])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}
			n.ToPersonID = ""

			Expect(c.CreateMessage(&n)).To(Equal(messages.Items[1]))
		})

		It("fails if a nil argument is provided", func() {
			p, err := c.CreateMessage(nil)
			Expect(err).To(MatchError("nil message"))
			Expect(p).To(BeNil())
		})

		It("fails if room ID, person ID, *and* person email are all empty", func() {
			n.RoomID = ""
			n.ToPersonEmail = ""
			n.ToPersonID = ""

			p, err := c.CreateMessage(&n)
			Expect(err).To(MatchError("message requires a room ID, person ID, or email to send to"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.CreateMessage(&n)
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("DeleteMessage", func() {
		It("deletes a message", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/%s", MessagesURL, messages.Items[0].ID)))
				Expect(req.Method).To(Equal("DELETE"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				r := &http.Response{
					Body:       closer(&bytes.Buffer{}), // empty body
					StatusCode: http.StatusNoContent,    // deletes are weird and return 204 instead of 200
				}
				return r, nil
			}

			Expect(c.DeleteMessage(messages.Items[0].ID)).To(Succeed())
		})

		It("fails if the message ID is empty", func() {
			Expect(c.DeleteMessage("")).To(MatchError("no message ID specified"))
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			Expect(c.DeleteMessage("1")).To(MatchError(mockErr))
		})
	})
})
