package spark

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// The bars don't line up here for most fonts, but they do in a terminal
const warning = "\n" +
	"\x1b[31m╔═══ WARNING ══════════════════════════════════════════════════════════════════════════════════════════════════════════╗\n" +
	"\x1b[31m║ \x1b[0mThe spark package includes tests that connect to the actual Spark endpoints for both GET and POST requests. To run   \x1b[31m║\n" +
	"\x1b[31m║ \x1b[0mthese tests, please set the '\x1b[36mSPARK_TOKEN\x1b[0m' environment variable to a valid Spark authentication token.                \x1b[31m║\n" +
	"\x1b[31m║ \x1b[0mRunning only mock tests on this runthrough...                                                                        \x1b[31m║\n" +
	"\x1b[31m╚══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝\x1b[0m\n\n"

const pkg = "Spark"

var token string

var mockErr = fmt.Errorf("mock error")

func TestReporter(t *testing.T) {
	if testing.Short() {
		return
	}

	if t, set := os.LookupEnv("SPARK_TOKEN"); set {
		token = t
	} else {
		fmt.Printf(warning)
	}

	RegisterFailHandler(Fail)

	RunSpecs(t, pkg+" Package")
}

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (r *http.Response, err error) {
	if m.DoFunc != nil {
		r, err = m.DoFunc(req)
	}
	return
}

// Wrapper for an io.Reader to give it a no-op Close method, so it can fulfill the io.ReadCloser interface.
// This is mostly because a bytes.Buffer does not have a Close method, but http.Response bodies need to have a close
// method.  This is functionally identical to an ioutil.NopCloser, except this version allows tests to *check* whether
// close was called or not.
type readCloser struct {
	io.Reader
	closed bool
}

func (r *readCloser) Close() error {
	r.closed = true
	return nil
}

func closer(r io.Reader) *readCloser {
	return &readCloser{Reader: r}
}

type failReader struct{}

func (*failReader) Read([]byte) (int, error) {
	return 0, mockErr
}
