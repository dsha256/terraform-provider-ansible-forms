package httpclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/exp/slog"
)

// Request represents a request to a REST API
type Request struct {
	Method string         `json:"method"`
	Body   map[string]any `json:"body"`
	Query  url.Values     `json:"query"`
	// uuid   string
}

// BuildHTTPReq builds an HTTP request to carry out the REST request
func (r *Request) BuildHTTPReq(c *HTTPClient, baseURL string) (*http.Request, error) {
	_url, err := r.BuildURL(c, baseURL, "")
	if err != nil {
		return nil, err
	}
	var req *http.Request
	var body io.Reader
	if len(r.Body) != 0 {
		var bodyJSON []byte
		bodyJSON, err = json.Marshal(r.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(bodyJSON)
	}
	req, err = http.NewRequest(r.Method, _url, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	//req.SetBasicAuth(c.cxProfile.Username, c.cxProfile.Password)

	token, err := r.getToken(c)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// telemetry header
	req.Header.Set("X-Dot-Client-App", c.tag)
	// TODO: low pty: add support for form data (require to create a file)

	return req, err
}

// BuildURL using Host, ApiRoot, baseURL, uuid, any query element
func (r *Request) BuildURL(c *HTTPClient, baseURL string, uuid string) (string, error) {
	var err error
	if c == nil {
		err = errors.New("error in BuildUrl, HTTPClient is nil")
	} else if r == nil {
		err = errors.New("error in BuildUrl, request is nil")
	} else if c.cxProfile.Hostname == "" || c.cxProfile.APIRoot == "" {
		err = errors.New("error in BuildUrl, Hostname and APIRoot are required")
	}
	if err != nil {
		return "", err
	}
	u := &url.URL{
		Scheme: "https",
		Host:   c.cxProfile.Hostname,
		Path:   c.cxProfile.APIRoot,
	}
	u = u.JoinPath(baseURL, uuid)
	if len(r.Query) != 0 {
		u.RawQuery = r.Query.Encode()
	}

	return u.String(), nil
}

type authResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (r *Request) getToken(c *HTTPClient) (string, error) {
	_url, err := r.BuildURL(c, "auth/login", "")
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, _url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.cxProfile.Username, c.cxProfile.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			slog.Error("error closing body", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var authResp authResponse
	if err = json.Unmarshal(body, &authResp); err != nil {
		return "", err
	}

	return authResp.Token, nil
}
