package alibaba

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultBaseURL = "https://gw.open.1688.com/openapi"
	DefaultTimeout = 30 * time.Second
	MaxRetries     = 3
)

// Client is the 1688 OpenAPI client.
type Client struct {
	BaseURL    string
	AppKey     string
	AppSecret  string
	httpClient *http.Client
}

// ApiResponse wraps 1688 API common response fields.
type ApiResponse struct {
	Success bool            `json:"success"`
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// NewClient creates a new 1688 API client.
func NewClient(appKey, appSecret string) *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		AppKey:     appKey,
		AppSecret:  appSecret,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

func (c *Client) buildSignedURL(path string, params url.Values) string {
	if params == nil {
		params = url.Values{}
	}
	params.Set("app_key", c.AppKey)
	ts := time.Now().Format("2006-01-02 15:04:05")
	params.Set("timestamp", ts)
	params.Set("format", "json")
	params.Set("v", "2.0")
	params.Set("sign_method", "sha256")

	sign := generateSign(params, c.AppSecret)
	params.Set("sign", sign)

	return c.BaseURL + path + "?" + params.Encode()
}

// Get sends a signed GET request to 1688 API with retry.
func (c *Client) Get(path string, params url.Values) (*ApiResponse, error) {
	reqURL := c.buildSignedURL(path, params)

	var lastErr error
	for i := 0; i < MaxRetries; i++ {
		resp, err := c.httpClient.Get(reqURL)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		var apiResp ApiResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			lastErr = err
			continue
		}

		return &apiResp, nil
	}
	return nil, fmt.Errorf("alibaba api failed after %d retries: %w", MaxRetries, lastErr)
}

// Post sends a signed POST request to 1688 API with retry.
func (c *Client) Post(path string, params url.Values, body io.Reader) (*ApiResponse, error) {
	reqURL := c.buildSignedURL(path, params)

	var lastErr error
	for i := 0; i < MaxRetries; i++ {
		resp, err := c.httpClient.Post(reqURL, "application/x-www-form-urlencoded", body)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		var apiResp ApiResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			lastErr = err
			continue
		}

		return &apiResp, nil
	}
	return nil, fmt.Errorf("alibaba api failed after %d retries: %w", MaxRetries, lastErr)
}
