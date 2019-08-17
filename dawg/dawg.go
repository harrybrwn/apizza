package dawg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultLang is the package language variable
	DefaultLang = "en"

	host = "order.dominos.com"
)

var cli = &http.Client{
	Timeout: time.Duration(10 * time.Second),
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func get(path string, params URLParam) ([]byte, error) {
	if params == nil {
		params = &Params{}
	}
	return send(&http.Request{
		Method: "GET",
		Host:   host,
		Proto:  "HTTP/1.1",
		Header: make(http.Header),
		URL: &url.URL{
			Scheme:   "https",
			Host:     host,
			Path:     path,
			RawQuery: params.Encode(),
		},
	})
}

func post(path string, data []byte) ([]byte, error) {
	return send(&http.Request{
		Method: "POST",
		Host:   host,
		Proto:  "HTTP/1.1",
		Body:   ioutil.NopCloser(bytes.NewReader(data)),
		Header: make(http.Header),
		URL: &url.URL{
			Scheme: "https",
			Host:   host,
			Path:   path,
		},
	})
}

func send(req *http.Request) ([]byte, error) {
	var buf bytes.Buffer

	req.Header.Add(
		"User-Agent",
		"Dominos API Wrapper for GO - "+time.Now().String(),
	)

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return buf.Bytes(), fmt.Errorf("bad response code: %d", resp.StatusCode)
	}
	return buf.Bytes(), nil
}
