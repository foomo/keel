package keeltest

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
)

type (
	HTTPClient struct {
		http.Client
		baseURL string
	}
	HTTPClientOption func(c *HTTPClient)
)

func HTTPClientWithCookieJar(v *cookiejar.Jar) HTTPClientOption {
	return func(c *HTTPClient) {
		c.Client.Jar = v
	}
}

func HTTPClientWithBaseURL(v string) HTTPClientOption {
	return func(c *HTTPClient) {
		c.baseURL = v
	}
}

func NewHTTPClient(opts ...HTTPClientOption) *HTTPClient {
	inst := &HTTPClient{
		Client:  http.Client{},
		baseURL: "",
	}

	for _, opt := range opts {
		opt(inst)
	}

	return inst
}

func (c *HTTPClient) Get(ctx context.Context, path string) ([]byte, int, error) {
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil); err != nil {
		return nil, 0, err
	} else if resp, err := c.Client.Do(req); err != nil {
		return nil, 0, err
	} else if body, err := c.readBody(resp); err != nil {
		return nil, 0, err
	} else {
		return body, resp.StatusCode, nil
	}
}

func (c *HTTPClient) Post(ctx context.Context, path string, data interface{}) ([]byte, int, error) {
	var req *http.Request
	if v, err := json.Marshal(data); err != nil {
		return nil, 0, err
	} else if r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewBuffer(v)); err != nil {
		return nil, 0, err
	} else {
		req = r
	}
	req.Header.Set("Content-Type", "application/json")
	if resp, err := c.Client.Do(req); err != nil {
		return nil, 0, err
	} else if body, err := c.readBody(resp); err != nil {
		return nil, 0, err
	} else {
		return body, resp.StatusCode, nil
	}
}

func (c *HTTPClient) readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		return body, nil
	}
}
