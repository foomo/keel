package keeltest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
)

type (
	HTTPClient struct {
		http.Client
		BaseURL string
	}
	HTTPClientOption func(c *HTTPClient)
)

func HTTPClientWithCookieJar(v *cookiejar.Jar) HTTPClientOption {
	return func(c *HTTPClient) {
		c.Jar = v
	}
}

func HTTPClientWithBaseURL(v string) HTTPClientOption {
	return func(c *HTTPClient) {
		c.BaseURL = v
	}
}

func NewHTTPClient(opts ...HTTPClientOption) *HTTPClient {
	inst := &HTTPClient{
		Client:  http.Client{},
		BaseURL: "",
	}

	for _, opt := range opts {
		opt(inst)
	}

	return inst
}

func (c *HTTPClient) Get(ctx context.Context, path string) ([]byte, int, error) {
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil); err != nil {
		return nil, 0, err
	} else if resp, err := c.Do(req); err != nil {
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
	} else if r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewBuffer(v)); err != nil {
		return nil, 0, err
	} else {
		req = r
	}

	req.Header.Set("Content-Type", "application/json")

	if resp, err := c.Do(req); err != nil {
		return nil, 0, err
	} else if body, err := c.readBody(resp); err != nil {
		return nil, 0, err
	} else {
		return body, resp.StatusCode, nil
	}
}

func (c *HTTPClient) readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	if body, err := io.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		return body, nil
	}
}
