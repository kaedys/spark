package spark

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("API", func() {
	var (
		c       *client
		mockCli *mockHTTPClient
		u       = "mock.url.com/mock"
		body    = []byte("mock body")
	)

	BeforeEach(func() {
		c = New("mock").(*client)
		mockCli = new(mockHTTPClient)
		httpCli = mockCli // set client global to a mock
	})

	Describe("request", func() {
		It("sets the expected headers", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(u))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json; charset=utf-8"))

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			req, err := http.NewRequest("GET", u, nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := c.request(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(Equal(body))
		})

		It("calls Close() on the body", func() {
			cls := closer(bytes.NewBuffer(body))

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				r := &http.Response{
					Body:       cls,
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			req, err := http.NewRequest("GET", u, nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := c.request(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(cls.closed).To(BeTrue())
			Expect(resp).To(Equal(body))
		})

		It("doesn't close the body if Do() fails", func() {
			cls := closer(bytes.NewBuffer(body))

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				r := &http.Response{
					Body:       cls,
					StatusCode: http.StatusOK,
				}
				return r, mockErr
			}

			req, err := http.NewRequest("GET", u, nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := c.request(req)
			Expect(err).To(MatchError(mockErr))
			Expect(cls.closed).To(BeFalse())
			Expect(resp).To(BeEmpty())
		})

		It("handles a body read error properly", func() {
			cls := closer(&failReader{})

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				r := &http.Response{
					Body:       cls,
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			req, err := http.NewRequest("GET", u, nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := c.request(req)
			Expect(err).To(MatchError(mockErr))
			Expect(cls.closed).To(BeTrue())
			Expect(resp).To(BeEmpty())
		})

		It("handles an unexpected HTTP status code", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusInternalServerError,
				}
				return r, nil
			}

			req, err := http.NewRequest("GET", u, nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := c.request(req)
			Expect(err.Error()).To(ContainSubstring("HTTP Status 500"))
			Expect(resp).To(BeEmpty())
		})
	})

	Describe("getRequest", func() {
		It("calls with the correct method and values", func() {
			vals := map[string][]string{
				"1": {"a"},
				"2": {"b"},
				"3": {"c"},
			}

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(u))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json; charset=utf-8"))

				for k, v := range vals {
					Expect(req.URL.Query().Get(k)).To(Equal(v[0]))
				}

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			resp, err := c.getRequest(u, url.Values(vals))
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(Equal(body))
		})

		It("handles parameters that are already in the url", func() {
			vals := map[string][]string{
				"1": {"a"},
				"2": {"b"},
				"3": {"c"},
			}

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(u))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json; charset=utf-8"))

				for k, v := range vals {
					Expect(req.URL.Query().Get(k)).To(Equal(v[0]))
				}

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			u2 := fmt.Sprintf("%s?1=a&2=b&3=c", u)

			resp, err := c.getRequest(u2, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(Equal(body))
		})

		It("handles a NewRequest() error properly", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				// This shouldn't be called in this test.  If it is, fail the test
				Fail("unexpected call to http.Client.Do()")
				return nil, nil
			}

			u2 := ":123" // invalid URL
			resp, err := c.getRequest(u2, nil)
			Expect(err).To(MatchError(fmt.Sprintf("parse %s: missing protocol scheme", u2)))
			Expect(resp).To(BeEmpty())
		})
	})

	Describe("postRequest", func() {
		It("calls with the correct method and body", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(u))
				Expect(req.Method).To(Equal("POST"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json; charset=utf-8"))

				b, err := ioutil.ReadAll(req.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(b).To(Equal(body))

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			resp, err := c.postRequest(u, bytes.NewBuffer(body))
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(Equal(body))
		})

		It("handles a NewRequest() error properly", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				// This shouldn't be called in this test.  If it is, fail the test
				Fail("unexpected call to http.Client.Do()")
				return nil, nil
			}

			u2 := ":123" // invalid URL
			resp, err := c.postRequest(u2, nil)
			Expect(err).To(MatchError(fmt.Sprintf("parse %s: missing protocol scheme", u2)))
			Expect(resp).To(BeEmpty())
		})
	})

	Describe("putRequest", func() {
		It("calls with the correct method and body", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(u))
				Expect(req.Method).To(Equal("PUT"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json; charset=utf-8"))

				b, err := ioutil.ReadAll(req.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(b).To(Equal(body))

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			resp, err := c.putRequest(u, bytes.NewBuffer(body))
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(Equal(body))
		})

		It("handles a NewRequest() error properly", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				// This shouldn't be called in this test.  If it is, fail the test
				Fail("unexpected call to http.Client.Do()")
				return nil, nil
			}

			u2 := ":123" // invalid URL
			resp, err := c.putRequest(u2, nil)
			Expect(err).To(MatchError(fmt.Sprintf("parse %s: missing protocol scheme", u2)))
			Expect(resp).To(BeEmpty())
		})
	})

	Describe("deleteRequest", func() {
		It("calls with the correct method and body", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.String()).To(Equal(u))
				Expect(req.Method).To(Equal("DELETE"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json; charset=utf-8"))

				Expect(req.Body).To(BeNil())

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			resp, err := c.deleteRequest(u)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(Equal(body))
		})

		It("handles a NewRequest() error properly", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				// This shouldn't be called in this test.  If it is, fail the test
				Fail("unexpected call to http.Client.Do()")
				return nil, nil
			}

			u2 := ":123" // invalid URL
			resp, err := c.deleteRequest(u2)
			Expect(err).To(MatchError(fmt.Sprintf("parse %s: missing protocol scheme", u2)))
			Expect(resp).To(BeEmpty())
		})
	})

	Describe("getRequestWithPaging", func() {
		It("calls with the correct method and values", func() {
			vals := map[string][]string{
				"1": {"a"},
				"2": {"b"},
				"3": {"c"},
			}
			max := 50

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(u))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json; charset=utf-8"))

				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				for k, v := range vals {
					Expect(req.URL.Query().Get(k)).To(Equal(v[0]))
				}

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, url.Values(vals), max)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(ConsistOf([][]byte{body}))
		})

		It("handles parameters that are already in the url", func() {
			vals := map[string][]string{
				"1": {"a"},
				"2": {"b"},
				"3": {"c"},
			}
			max := 50

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				uri := strings.Split(req.URL.String(), "?")[0]
				Expect(uri).To(Equal(u))
				Expect(req.Method).To(Equal("GET"))
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer mock"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json; charset=utf-8"))

				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max)))
				for k, v := range vals {
					Expect(req.URL.Query().Get(k)).To(Equal(v[0]))
				}

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			u2 := fmt.Sprintf("%s?1=a&2=b&3=c", u)

			resp, err := c.getRequestWithPaging(u2, nil, max)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(ConsistOf([][]byte{body}))
		})

		It("calls with the client max if it's smaller than the max argument", func() {
			max := 50
			expected := 10
			c.pageMax = expected

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", expected)))
				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, nil, max)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(ConsistOf([][]byte{body}))
		})

		It("pages", func() {
			expectedCalls := 10
			calls := 0
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", c.pageMax)))

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
				}
				if calls++; calls < expectedCalls {
					r.Header = map[string][]string{
						"Link": {fmt.Sprintf("<%s>; rel=\"next\"", u)},
					}
				}

				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, nil, 0)
			Expect(err).ToNot(HaveOccurred())
			slice := make([][]byte, expectedCalls)
			for i := 0; i < expectedCalls; i++ {
				slice[i] = body
			}
			Expect(resp).To(ConsistOf(slice))
			Expect(calls).To(Equal(expectedCalls))
		})

		It("pages until max is hit", func() {
			max := 50
			clientmax := 10
			c.pageMax = clientmax
			expectedCalls := 5

			calls := 0
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				calls++
				Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", clientmax)))

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
					Header: map[string][]string{
						"Link": {fmt.Sprintf("<%s>; rel=\"next\"", u)},
					},
				}

				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, nil, max)
			Expect(err).ToNot(HaveOccurred())
			slice := make([][]byte, expectedCalls)
			for i := 0; i < expectedCalls; i++ {
				slice[i] = body
			}
			Expect(resp).To(ConsistOf(slice))
			Expect(calls).To(Equal(expectedCalls))
		})

		It("pages with a non-integer multiple of client max", func() {
			max := 55
			clientmax := 10
			c.pageMax = clientmax
			expectedCalls := 6

			calls := 0
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				if calls++; calls < expectedCalls {
					Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", clientmax)))
				} else {
					Expect(req.URL.Query().Get("max")).To(Equal(fmt.Sprintf("%d", max%clientmax)))
				}

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
					Header: map[string][]string{
						"Link": {fmt.Sprintf("<%s>; rel=\"next\"", u)},
					},
				}

				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, nil, max)
			Expect(err).ToNot(HaveOccurred())
			slice := make([][]byte, expectedCalls)
			for i := 0; i < expectedCalls; i++ {
				slice[i] = body
			}
			Expect(resp).To(ConsistOf(slice))
			Expect(calls).To(Equal(expectedCalls))
		})

		It("calls Close() on the body", func() {
			cls := closer(bytes.NewBuffer(body))

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				r := &http.Response{
					Body:       cls,
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, nil, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(cls.closed).To(BeTrue())
			Expect(resp).To(ConsistOf([][]byte{body}))
		})

		It("calls Close() on the body on each iteration", func() {
			cls1 := closer(bytes.NewBuffer(body))
			cls2 := closer(bytes.NewBuffer(body))

			called := false
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				var r *http.Response
				if !called {
					called = true
					r = &http.Response{
						Body:       cls1,
						StatusCode: http.StatusOK,
						Header: map[string][]string{
							"Link": {fmt.Sprintf("<%s>; rel=\"next\"", u)},
						},
					}
				} else {
					r = &http.Response{
						Body:       cls2,
						StatusCode: http.StatusOK,
					}
				}
				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, nil, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(cls1.closed).To(BeTrue())
			Expect(cls2.closed).To(BeTrue())
			Expect(resp).To(ConsistOf([][]byte{body, body}))
		})

		It("doesn't close the body if Do() fails", func() {
			cls := closer(bytes.NewBuffer(body))

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				r := &http.Response{
					Body:       cls,
					StatusCode: http.StatusOK,
				}
				return r, mockErr
			}

			resp, err := c.getRequestWithPaging(u, nil, 0)
			Expect(err).To(MatchError(mockErr))
			Expect(cls.closed).To(BeFalse())
			Expect(resp).To(BeEmpty())
		})

		It("handles a NewRequest() error properly", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				// This shouldn't be called in this test.  If it is, fail the test
				Fail("unexpected call to http.Client.Do()")
				return nil, nil
			}

			u2 := ":123" // invalid URL
			resp, err := c.getRequestWithPaging(u2, nil, 0)
			Expect(err).To(MatchError(fmt.Sprintf("parse %s: missing protocol scheme", u2)))
			Expect(resp).To(BeEmpty())
		})

		It("handles a NewRequest() error properly after the first iteration", func() {
			called := false
			u2 := ":123" // invalid URL

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				if called {
					// should only be called once, fail if called a second time
					Fail("unexpected call to http.Client.Do()")
				}
				called = true

				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusOK,
					Header: map[string][]string{
						"Link": {fmt.Sprintf("<%s>; rel=\"next\"", u2)}, // invalid URL
					},
				}
				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, nil, 0)
			Expect(err).To(MatchError(fmt.Sprintf("parse %s: missing protocol scheme", u2)))
			Expect(resp).To(ConsistOf([][]byte{body}))
		})

		It("handles a body read error properly", func() {
			cls := closer(&failReader{})

			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				r := &http.Response{
					Body:       cls,
					StatusCode: http.StatusOK,
				}
				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, nil, 0)
			Expect(err).To(MatchError(mockErr))
			Expect(cls.closed).To(BeTrue())
			Expect(resp).To(BeEmpty())
		})

		It("handles a body read error properly after the first iteration", func() {
			cls1 := closer(bytes.NewBuffer(body))
			cls2 := closer(&failReader{})

			called := false
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				var r *http.Response
				if !called {
					called = true
					r = &http.Response{
						Body:       cls1,
						StatusCode: http.StatusOK,
						Header: map[string][]string{
							"Link": {fmt.Sprintf("<%s>; rel=\"next\"", u)},
						},
					}
				} else {
					r = &http.Response{
						Body:       cls2,
						StatusCode: http.StatusOK,
					}
				}
				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, nil, 0)
			Expect(err).To(MatchError(mockErr))
			Expect(cls1.closed).To(BeTrue())
			Expect(cls2.closed).To(BeTrue())
			Expect(resp).To(ConsistOf([][]byte{body}))
		})

		It("handles an unexpected HTTP status code", func() {
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				r := &http.Response{
					Body:       closer(bytes.NewBuffer(body)),
					StatusCode: http.StatusInternalServerError,
				}
				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, nil, 0)
			Expect(err.Error()).To(ContainSubstring("HTTP Status 500"))
			Expect(resp).To(BeEmpty())
		})

		It("handles an unexpected HTTP status code after the first iteration", func() {
			cls1 := closer(bytes.NewBuffer(body))
			cls2 := closer(bytes.NewBuffer(body))

			called := false
			mockCli.DoFunc = func(req *http.Request) (*http.Response, error) {
				var r *http.Response
				if !called {
					called = true
					r = &http.Response{
						Body:       cls1,
						StatusCode: http.StatusOK,
						Header: map[string][]string{
							"Link": {fmt.Sprintf("<%s>; rel=\"next\"", u)},
						},
					}
				} else {
					r = &http.Response{
						Body:       cls2,
						StatusCode: http.StatusInternalServerError,
					}
				}
				return r, nil
			}

			resp, err := c.getRequestWithPaging(u, nil, 0)
			Expect(err.Error()).To(ContainSubstring("HTTP Status 500"))
			Expect(resp).To(ConsistOf([][]byte{body}))
		})
	})
})
