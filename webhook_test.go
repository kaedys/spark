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

var _ = Describe("Webhook (Mock)", func() {
	var c Client
	var mockCli *mockHTTPClient

	var webhooks WebhookList

	BeforeEach(func() {
		c = New("mock")
		mockCli = new(mockHTTPClient)
		httpCli = mockCli // set client global to a mock

		webhooks = WebhookList{
			Items: []*Webhook{
				{
					ID:        "1",
					Name:      "webhook 1",
					TargetURL: "url 1",
					Resource:  "resource 1",
					Event:     "event 1",
				},
				{
					ID:        "2",
					Name:      "webhook 2",
					TargetURL: "url 2",
					Resource:  "resource 2",
					Event:     "event 2",
				},
				{
					ID:        "3",
					Name:      "webhook 3",
					TargetURL: "url 3",
					Resource:  "resource 3",
					Event:     "event 3",
				},
			},
		}
	})

	Describe("GetWebhook", func() {
		It("gets a webhook by ID", func() {
			webhookID := webhooks.Items[0].ID

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/%s", WebhooksURL, webhookID)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(webhooks.Items[0])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.GetWebhook(webhookID)).To(Equal(webhooks.Items[0]))
		})

		It("fails if no webhook ID is specified", func() {
			p, err := c.GetWebhook("")
			Expect(err).To(MatchError("no webhook ID specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.GetWebhook("1")
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("ListWebhooks", func() {
		It("gets a list of webhooks", func() {
			max := len(webhooks.Items)

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(WebhooksURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(webhooks)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListWebhooks(max)).To(ConsistOf(webhooks.Items))
		})

		It("gets a list of webhooks with a maximum", func() {
			max := len(webhooks.Items) - 1

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(WebhooksURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				webhooks.Items = webhooks.Items[:max]

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(webhooks)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListWebhooks(max)).To(ConsistOf(webhooks.Items))
		})

		It("sets max parameter to the client max if max arg = 0", func() {
			max := 0
			cmax := 25
			c = c.SetMaxPerPage(cmax)

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(WebhooksURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(webhooks)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.ListWebhooks(max)).To(ConsistOf(webhooks.Items))
		})

		It("pages if max > client max", func() {
			max := len(webhooks.Items)
			cmax := 1
			c = c.SetMaxPerPage(cmax)

			calls := 0
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(WebhooksURL))
				if calls == 0 {
					Expect(req.URL.Query().Get("after")).To(BeEmpty())
				} else {
					Expect(req.URL.Query().Get("after")).To(Equal(webhooks.Items[calls-1].ID))
				}
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				p := WebhookList{
					Items: webhooks.Items[calls : calls+1],
				}

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(p)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}

				if calls < max {
					r.Header = map[string][]string{
						"Link": {fmt.Sprintf("<%s?max=%d&after=%s>; rel=\"next\"", WebhooksURL, cmax, webhooks.Items[calls].ID)},
					}
				}

				calls++

				return r, nil
			}

			Expect(c.ListWebhooks(max)).To(ConsistOf(webhooks.Items))
			Expect(calls).To(BeEquivalentTo(3))
		})

		It("pages until it stops getting next links if max = 0", func() {
			max := 0
			cmax := len(webhooks.Items)
			c = c.SetMaxPerPage(cmax)

			calls := 0
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(WebhooksURL))
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", cmax)))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(webhooks)).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				calls++

				if calls < 10 {
					r.Header = map[string][]string{
						"Link": {fmt.Sprintf("<%s>; rel=\"next\"", WebhooksURL)},
					}
				}

				return r, nil
			}

			Expect(c.ListWebhooks(max)).To(HaveLen(len(webhooks.Items) * 10))
			Expect(calls).To(BeEquivalentTo(10))
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.ListWebhooks(0)
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("CreateWebhook", func() {
		var n NewWebhook

		BeforeEach(func() {
			// Wash it through the json package, because honestly that's the easiest way to copy a struct to
			// a struct with a subset of the same fields
			var b bytes.Buffer
			Expect(json.NewEncoder(&b).Encode(webhooks.Items[0])).To(Succeed())
			Expect(json.NewDecoder(&b).Decode(&n)).To(Succeed())

			Expect(n.Name).To(Equal(webhooks.Items[0].Name))
			Expect(n.TargetURL).To(Equal(webhooks.Items[0].TargetURL))
			Expect(n.Resource).To(Equal(webhooks.Items[0].Resource))
			Expect(n.Event).To(Equal(webhooks.Items[0].Event))
			Expect(n.Filter).To(Equal(webhooks.Items[0].Filter))
			Expect(n.Secret).To(Equal(webhooks.Items[0].Secret))
		})

		It("creates a webhook", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(WebhooksURL))
				Expect(req.Method).To(Equal("POST"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var p NewWebhook
				Expect(json.NewDecoder(req.Body).Decode(&p)).To(Succeed())
				Expect(p).To(Equal(n))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(webhooks.Items[1])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.CreateWebhook(&n)).To(Equal(webhooks.Items[1]))
		})

		It("fails if a nil argument is provided", func() {
			p, err := c.CreateWebhook(nil)
			Expect(err).To(MatchError("nil webhook"))
			Expect(p).To(BeNil())
		})

		It("fails if no webhook name is provided", func() {
			n.Name = ""

			p, err := c.CreateWebhook(&n)
			Expect(err).To(MatchError("no webhook name specified"))
			Expect(p).To(BeNil())
		})

		It("fails if no webhook target URL is provided", func() {
			n.TargetURL = ""

			p, err := c.CreateWebhook(&n)
			Expect(err).To(MatchError("no webhook target URL specified"))
			Expect(p).To(BeNil())
		})

		It("fails if no webhook resource is provided", func() {
			n.Resource = ""

			p, err := c.CreateWebhook(&n)
			Expect(err).To(MatchError("no webhook resource specified"))
			Expect(p).To(BeNil())
		})

		It("fails if no webhook event is provided", func() {
			n.Event = ""

			p, err := c.CreateWebhook(&n)
			Expect(err).To(MatchError("no webhook event specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.CreateWebhook(&n)
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("UpdateWebhook", func() {
		It("updates a webhook", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/%s", WebhooksURL, webhooks.Items[0].ID)))
				Expect(req.Method).To(Equal("PUT"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				var p Webhook
				Expect(json.NewDecoder(req.Body).Decode(&p)).To(Succeed())
				Expect(&p).To(Equal(webhooks.Items[0]))

				var b bytes.Buffer
				Expect(json.NewEncoder(&b).Encode(webhooks.Items[1])).To(Succeed())
				r := &http.Response{
					Body:       closer(&b),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			Expect(c.UpdateWebhook(webhooks.Items[0])).To(Equal(webhooks.Items[1]))
		})

		It("fails if a nil argument is provided", func() {
			p, err := c.UpdateWebhook(nil)
			Expect(err).To(MatchError("nil webhook"))
			Expect(p).To(BeNil())
		})

		It("fails if an empty webhook ID is provided", func() {
			webhooks.Items[0].ID = ""
			p, err := c.UpdateWebhook(webhooks.Items[0])
			Expect(err).To(MatchError("no webhook ID specified"))
			Expect(p).To(BeNil())
		})

		It("fails if an empty webhook name is provided", func() {
			webhooks.Items[0].Name = ""
			p, err := c.UpdateWebhook(webhooks.Items[0])
			Expect(err).To(MatchError("no webhook name specified"))
			Expect(p).To(BeNil())
		})

		It("fails if an empty webhook target URL is provided", func() {
			webhooks.Items[0].TargetURL = ""
			p, err := c.UpdateWebhook(webhooks.Items[0])
			Expect(err).To(MatchError("no webhook target URL specified"))
			Expect(p).To(BeNil())
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			p, err := c.UpdateWebhook(webhooks.Items[0])
			Expect(err).To(MatchError(mockErr))
			Expect(p).To(BeNil())
		})
	})

	Describe("DeleteWebhoook", func() {
		It("deletes a webhook", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(fmt.Sprintf("%s/%s", WebhooksURL, webhooks.Items[0].ID)))
				Expect(req.Method).To(Equal("DELETE"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))

				r := &http.Response{
					Body:       closer(&bytes.Buffer{}), // empty body
					StatusCode: http.StatusNoContent,    // deletes are weird and return 204 instead of 200
				}
				return r, nil
			}

			Expect(c.DeleteWebhook(webhooks.Items[0].ID)).To(Succeed())
		})

		It("fails if the webhook ID is empty", func() {
			Expect(c.DeleteWebhook("")).To(MatchError("no webhook ID specified"))
		})

		It("passes through errors encountered during the request", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			}
			Expect(c.DeleteWebhook("1")).To(MatchError(mockErr))
		})
	})
})
