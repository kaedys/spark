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

var _ = Describe("Person (Live)", func() {
	var c Client

	BeforeEach(func() {
		httpCli = httpClient(new(http.Client))

		if token == "" {
			Skip("token not set, skipping live tests")
		}
		c = New(token)
	})

	Describe("GetPerson", func() {
		It("gets myself", func() {
			Expect(c.GetMyself()).ToNot(BeNil())
		})

		It("gets myself by ID", func() {
			me, err := c.GetMyself()
			Expect(err).ToNot(HaveOccurred())
			Expect(me).ToNot(BeNil())

			me2, err := c.GetPerson(me.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(me2).ToNot(BeNil())
			Expect(me).To(Equal(me2))
		})
	})

	Describe("ListPeople", func() {
		It("gets a list of people", func() {
			me, err := c.GetMyself()
			Expect(err).ToNot(HaveOccurred())
			Expect(me).ToNot(BeNil())

			Expect(c.ListPeople(5, &PeopleListParams{Email: me.Emails[0]})).To(ConsistOf(me))
		})
	})
})

var _ = Describe("Person (Mock)", func() {
	var c Client
	var mockCli *mockHTTPClient

	var people People

	BeforeEach(func() {
		c = New("mock")
		mockCli = new(mockHTTPClient)
		httpCli = mockCli // set client global to a mock

		people = People{
			Items: []*Person{
				{
					ID:          "1",
					Emails:      []string{"hello1@world.com"},
					DisplayName: "test 1",
				},
				{
					ID:          "2",
					Emails:      []string{"hello2@world.com"},
					DisplayName: "test 2",
				},
				{
					ID:          "3",
					Emails:      []string{"hello3@world.com"},
					DisplayName: "test 3",
				},
			},
		}
	})

	Describe("GetPerson", func() {
		It("gets the 'me' person", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/me", PeopleURL)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(people.Items[0])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.GetMyself()).To(Equal(people.Items[0]))
		})

		It("gets a person by ID", func() {
			personID := people.Items[0].ID

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/%s", PeopleURL, personID)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(people.Items[0])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.GetPerson(personID)).To(Equal(people.Items[0]))
		})

		It("fails if no person ID is specified", func() {
			p, err := c.GetPerson("")
			Expect(err).To(MatchError("no person ID specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.GetPerson("1")
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("ListPeople", func() {
		It("gets a list of people", func() {
			max := len(people.Items)

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(PeopleURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(people)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListPeople(max, nil)).To(ConsistOf(people.Items))
		})

		It("gets a list of people with a maximum", func() {
			max := len(people.Items) - 1

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(PeopleURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				people.Items = people.Items[:max]

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(people)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListPeople(max, nil)).To(ConsistOf(people.Items))
		})

		It("sets max parameter to the client max if max arg = 0", func() {
			max := 0
			cmax := 25
			c = c.SetMaxPerPage(cmax)

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(PeopleURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(people)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListPeople(max, nil)).To(ConsistOf(people.Items))
		})

		It("pages if max > client max", func() {
			max := len(people.Items)
			cmax := 1
			c = c.SetMaxPerPage(cmax)

			calls := 0
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(PeopleURL))
				if calls == 0 {
					Expect(req.URL.Query().Get("after")).To(BeEmpty())
				} else {
					Expect(req.URL.Query().Get("after")).To(Equal(people.Items[calls-1].ID))
				}
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				p := People{
					Items: people.Items[calls : calls+1],
				}

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(p)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}

				if calls < max {
					r.Header = map[string][]string{
						"Link": {fmt.Sprintf("<%s?max=%d&after=%s>; rel=\"next\"", PeopleURL, cmax, people.Items[calls].ID)},
					}
				}

				calls++

				return r, nil
			}

			Expect(c.ListPeople(max, nil)).To(ConsistOf(people.Items))
			Expect(calls).To(BeEquivalentTo(3))
		})

		It("pages until it stops getting next links if max = 0", func() {
			max := 0
			cmax := len(people.Items)
			c = c.SetMaxPerPage(cmax)

			calls := 0
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(PeopleURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(people)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				calls++

				if calls < 10 {
					r.Header = map[string][]string{
						"Link": {fmt.Sprintf("<%s>; rel=\"next\"", PeopleURL)},
					}
				}

				return r, nil
			}

			Expect(c.ListPeople(max, nil)).To(HaveLen(len(people.Items) * 10))
			Expect(calls).To(BeEquivalentTo(10))
		})

		It("applies a parameter list", func() {
			max := len(people.Items)
			params := PeopleListParams{
				Email:       "test email",
				DisplayName: "test name",
				ID:          "test ID",
				OrgID:       "test org ID",
			}

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(PeopleURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				for k, v := range params.values() {
					Expect(req.URL.Query().Get(k)).To(Equal(v[0]), fmt.Sprintf("MISSING [%s] %+v", k, req.URL.Query()))
				}

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(people)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListPeople(max, &params)).To(ConsistOf(people.Items))
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.ListPeople(0, nil)
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("CreatePerson", func() {
		It("creates a person", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(PeopleURL))
				Expect(req.Method).To(Equal("POST"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var p Person
				Expect(json.NewDecoder(req.Body).Decode(&p)).To(Succeed())
				Expect(&p).To(Equal(people.Items[0]))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(people.Items[1])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.CreatePerson(people.Items[0])).To(Equal(people.Items[1]))
		})

		It("fails if a nil argument is provided", func() {
			p, err := c.CreatePerson(nil)
			Expect(err).To(MatchError("nil person"))
			Expect(p).To(BeNil())
		})

		It("fails if the person has no emails", func() {
			people.Items[0].Emails = nil
			p, err := c.CreatePerson(people.Items[0])
			Expect(err).To(MatchError("no email specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.CreatePerson(people.Items[0])
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("UpdatePerson", func() {
		It("updates a person", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/%s", PeopleURL, people.Items[0].ID)))
				Expect(req.Method).To(Equal("PUT"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var p Person
				Expect(json.NewDecoder(req.Body).Decode(&p)).To(Succeed())
				Expect(&p).To(Equal(people.Items[0]))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(people.Items[1])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.UpdatePerson(people.Items[0])).To(Equal(people.Items[1]))
		})

		It("fails if a nil argument is provided", func() {
			p, err := c.UpdatePerson(nil)
			Expect(err).To(MatchError("nil person"))
			Expect(p).To(BeNil())
		})

		It("fails if the person has no ID", func() {
			people.Items[0].ID = ""
			p, err := c.UpdatePerson(people.Items[0])
			Expect(err).To(MatchError("no person ID specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.UpdatePerson(people.Items[0])
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("DeletePerson", func() {
		It("deletes a person", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/%s", PeopleURL, people.Items[0].ID)))
				Expect(req.Method).To(Equal("DELETE"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				r := &http.Response{
					Body:       closer(&bytes.Buffer{}), // empty body
					StatusCode: http.StatusNoContent,    // deletes are weird and return 204 instead of 200
				}
				return r, nil
			}

			Expect(c.DeletePerson(people.Items[0].ID)).To(Succeed())
		})

		It("fails if the person ID is empty", func() {
			Expect(c.DeletePerson("")).To(MatchError("no person ID specified"))
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			Expect(c.DeletePerson("1")).To(MatchError(mockErr))
		})
	})
})
