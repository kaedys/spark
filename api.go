package spark

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var httpCli = httpClient(new(http.Client))

func (c *client) request(req *http.Request) ([]byte, error) {
	// All requests require these headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// return code should be 200, or 204 for delete methods
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("HTTP Status %d: %q", res.StatusCode, string(bs))
	}

	return bs, nil
}

func (c *client) getRequest(url string, uv url.Values) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	params := req.URL.Query()
	for k, vals := range uv {
		for _, v := range vals {
			params.Add(k, v)
		}
	}
	req.URL.RawQuery = params.Encode()
	return c.request(req)
}

func (c *client) postRequest(url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	return c.request(req)
}

func (c *client) putRequest(url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	return c.request(req)
}

func (c *client) deleteRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	return c.request(req)
}

// Works like getRequest, except it handles paginated results.  It will retrieve up to max total entries, across
// however many pages are necessary, unless the server indicates that it is out of results before that point is reached.
// As long as the first page query  succeeds, this function will return any partial results it has successfully
// received even in the case of an error (ex. if it encounters an error retrieving page 3, pages 1 and 2 will still be
// returned).  As a special case, if max is set to 0, this function will retrieve *all* values that the server makes
// available.
func (c *client) getRequestWithPaging(uri string, uv url.Values, max int) ([][]byte, error) {
	all := false
	if max == 0 {
		all = true
	}

	var ret [][]byte
	for all || max > 0 {
		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			return ret, err
		}

		params := req.URL.Query()
		for k, vals := range uv {
			for _, v := range vals {
				params.Add(k, v)
			}
		}
		// We unconditionally overwrite the "max" parameter here.  We do this just in case the input uri has it
		// set, and also because the "next" urls returned by paged queries have max set, but we sometimes want
		// a different value that it sets for us.
		if all || max > c.pageMax {
			params["max"] = []string{fmt.Sprintf("%d", c.pageMax)}
		} else {
			params["max"] = []string{fmt.Sprintf("%d", max)}
		}

		// if max < pageMax, it'll go negative, but that'll end the loop just as effectively as setting it to 0.
		// If All is set, in theory this could overflow, but that would require receiving more than 2.1 billion values
		// (32-bit system) or 9 quintillion values (64-bit system), and if All is set, it doesn't really matter if it
		// overflows, because we're looping until we run out anyway. Fortunately, overflowing an int in Go is not an
		// error, it simply wraps around to positive integers.
		max -= c.pageMax

		req.URL.RawQuery = params.Encode()

		// All requests require these headers
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		res, err := httpCli.Do(req)
		if err != nil {
			return ret, err
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return ret, err
		}

		// Return code should be 200, or 204 for delete methods
		if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
			return ret, fmt.Errorf("HTTP Status %d: %q", res.StatusCode, res.Status)
		}

		ret = append(ret, b)

		// Check for pagination.  The Spark API indicates pagination by including a "Link" header.  This header
		// can contain multiple URLs, but the one we care about is the rel="next" one, as that URL will give us the
		// next page of results.  This will loop until the pagination stops or until the page limit argument is reached.
		// As a special case, if pageLimit == 0, this will loop until the server stops returning next URLs, regardless
		// of how many pages that involves.
		found := false

	headers:
		for k, v := range res.Header {
			if k == "Link" {
				for _, l := range v {
					spl := strings.Split(l, "; ")
					if spl[1] == `rel="next"` {
						found = true
						// The format of the header is `<url?params>; rel="next"`
						// The split above will leave spl[0] = `<url?params>`, so trim the first and last char
						uri = spl[0][1 : len(spl[0])-1]
					}
					break headers
				}
			}
		}
		if !found {
			// Ran out of next headers, break and return
			break
		}
	}
	return ret, nil
}
