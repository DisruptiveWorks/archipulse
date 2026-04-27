// Package cli implements the ArchiPulse command-line interface subcommands.
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Client is a thin HTTP wrapper that sends the session cookie on every request.
type Client struct {
	server string
	token  string
	http   *http.Client
}

// NewClient creates a Client targeting server with the given session token.
// token is the raw value of the ap_session cookie (empty = unauthenticated).
func NewClient(server, token string) *Client {
	return &Client{
		server: server,
		token:  token,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Do(method, path string, body any) (*http.Response, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
	}
	req, err := http.NewRequest(method, c.server+"/api/v1"+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Cookie", "ap_session="+c.token)
	}
	return c.http.Do(req)
}

// DoMultipart posts content as a multipart/form-data upload under the given field name.
func (c *Client) DoMultipart(path, field string, content io.Reader) (*http.Response, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile(field, "upload.xml")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(fw, content); err != nil {
		return nil, fmt.Errorf("write content: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close multipart: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.server+"/api/v1"+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Cookie", "ap_session="+c.token)
	}
	return c.http.Do(req)
}

// decode reads a JSON response body into out. Returns an error for non-2xx
// responses, extracting the "error" field from the JSON body when present.
func (c *Client) decode(resp *http.Response, out any) error {
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		if e.Error != "" {
			return fmt.Errorf("%s", e.Error)
		}
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
