package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/openmarkers/openmarkers-cli/internal/infrastructure/auth"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Store      *auth.Store
	Verbose    bool
	LogFunc    func(string, ...any)
}

func NewClient(baseURL string, store *auth.Store) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Store: store,
	}
}

func (c *Client) log(format string, args ...any) {
	if c.Verbose && c.LogFunc != nil {
		c.LogFunc(format, args...)
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	if c.Store != nil {
		tokens := c.Store.Load()
		if tokens != nil && tokens.ExpiresAt > 0 {
			if time.Now().Unix() >= tokens.ExpiresAt-30 {
				c.log("Token expired or expiring soon, refreshing...")
				if err := c.refreshToken(ctx, tokens); err != nil {
					return nil, fmt.Errorf("token refresh: %w", err)
				}
			}
		}
	}

	fullURL := c.BaseURL + path
	c.log("%s %s", method, fullURL)

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		c.log("Request body: %s", string(data))
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.Store != nil {
		tokens := c.Store.Load()
		if tokens != nil && tokens.AccessToken != "" {
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}

	c.log("Response: %d", resp.StatusCode)
	return resp, nil
}

func (c *Client) refreshToken(ctx context.Context, tokens *auth.Tokens) error {
	if tokens.RefreshToken == "" || tokens.TokenEndpoint == "" || tokens.ClientID == "" {
		return fmt.Errorf("missing refresh token or client credentials")
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {tokens.RefreshToken},
		"client_id":     {tokens.ClientID},
	}
	if tokens.ClientSecret != "" {
		form.Set("client_secret", tokens.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokens.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	tokens.AccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		tokens.RefreshToken = tokenResp.RefreshToken
	}
	tokens.ExpiresAt = time.Now().Unix() + tokenResp.ExpiresIn

	return c.Store.Save(tokens)
}

func decodeResponse(resp *http.Response, target any) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return NewAPIError(resp.StatusCode, errResp.Error)
		}
		return NewAPIError(resp.StatusCode, string(body))
	}

	if target != nil {
		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) Get(ctx context.Context, path string, target any) error {
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	return decodeResponse(resp, target)
}

func (c *Client) Post(ctx context.Context, path string, body any, target any) error {
	resp, err := c.doRequest(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	return decodeResponse(resp, target)
}

func (c *Client) Patch(ctx context.Context, path string, body any, target any) error {
	resp, err := c.doRequest(ctx, "PATCH", path, body)
	if err != nil {
		return err
	}
	return decodeResponse(resp, target)
}

func (c *Client) Delete(ctx context.Context, path string, target any) error {
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	return decodeResponse(resp, target)
}

func (c *Client) GetRaw(ctx context.Context, path string) ([]byte, error) {
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, NewAPIError(resp.StatusCode, errResp.Error)
		}
		return nil, NewAPIError(resp.StatusCode, string(body))
	}

	return body, nil
}
