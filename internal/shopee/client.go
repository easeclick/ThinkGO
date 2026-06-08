package shopee

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	DefaultBaseURL = "https://partner.shopeemobile.com"
	DefaultTimeout = 30 * time.Second
	MaxRetries     = 3
)

// Client is the Shopee OpenAPI client.
type Client struct {
	BaseURL    string
	PartnerID  int64
	PartnerKey string
	ShopID     int64
	httpClient *http.Client
}

// ApiResponse wraps Shopee API common response fields.
type ApiResponse struct {
	Error     string          `json:"error"`
	Message   string          `json:"message"`
	RequestID string          `json:"request_id"`
	RawData   json.RawMessage `json:"data,omitempty"`
}

// UnmarshalData parses the data field into the provided value.
func (r *ApiResponse) UnmarshalData(v interface{}) error {
	if r.RawData == nil {
		return nil
	}
	return json.Unmarshal(r.RawData, v)
}

// NewClient creates a new Shopee API client.
func NewClient(partnerID int64, partnerKey string, shopID int64) *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		PartnerID:  partnerID,
		PartnerKey: partnerKey,
		ShopID:     shopID,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

func (c *Client) buildSignedURL(path string, params map[string]string) string {
	if params == nil {
		params = make(map[string]string)
	}
	timestamp := time.Now().Unix()
	params["partner_id"] = strconv.FormatInt(c.PartnerID, 10)
	params["timestamp"] = strconv.FormatInt(timestamp, 10)

	if c.ShopID > 0 {
		params["shop_id"] = strconv.FormatInt(c.ShopID, 10)
	}

	sign := GenerateSign(params, c.PartnerKey)

	url := c.BaseURL + path + "?"
	for k, v := range params {
		url += k + "=" + v + "&"
	}
	url += "sign=" + sign
	return url
}

// Get sends a signed GET request to Shopee API with retry.
func (c *Client) Get(path string, params map[string]string) (*ApiResponse, error) {
	url := c.buildSignedURL(path, params)

	var lastErr error
	for i := 0; i < MaxRetries; i++ {
		resp, err := c.httpClient.Get(url)
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

		if apiResp.Error != "" {
			return nil, fmt.Errorf("shopee api error: %s - %s", apiResp.Error, apiResp.Message)
		}

		return &apiResp, nil
	}
	return nil, fmt.Errorf("shopee api failed after %d retries: %w", MaxRetries, lastErr)
}

// Post sends a signed POST request to Shopee API with retry.
func (c *Client) Post(path string, params map[string]string, body io.Reader) (*ApiResponse, error) {
	url := c.buildSignedURL(path, params)

	var lastErr error
	for i := 0; i < MaxRetries; i++ {
		resp, err := c.httpClient.Post(url, "application/json", body)
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

		if apiResp.Error != "" {
			return nil, fmt.Errorf("shopee api error: %s - %s", apiResp.Error, apiResp.Message)
		}

		return &apiResp, nil
	}
	return nil, fmt.Errorf("shopee api failed after %d retries: %w", MaxRetries, lastErr)
}
