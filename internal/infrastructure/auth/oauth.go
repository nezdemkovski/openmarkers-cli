package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type OAuthConfig struct {
	ServerURL string
	Store     *Store
	LogFunc   func(string, ...any)
}

type serverMetadata struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	RegistrationEndpoint  string `json:"registration_endpoint"`
}

type clientRegistration struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (o *OAuthConfig) log(format string, args ...any) {
	if o.LogFunc != nil {
		o.LogFunc(format, args...)
	}
}

func (o *OAuthConfig) Login(ctx context.Context) error {
	o.log("Discovering OAuth endpoints...")
	meta, err := o.discover(ctx)
	if err != nil {
		return fmt.Errorf("discover endpoints: %w", err)
	}

	tokens := o.Store.Load()
	var clientID, clientSecret string
	if tokens != nil && tokens.ClientID != "" {
		clientID = tokens.ClientID
		clientSecret = tokens.ClientSecret
	} else {
		o.log("Registering new OAuth client...")
		reg, err := o.register(ctx, meta.RegistrationEndpoint)
		if err != nil {
			return fmt.Errorf("client registration: %w", err)
		}
		clientID = reg.ClientID
		clientSecret = reg.ClientSecret
	}

	verifier, err := generateVerifier()
	if err != nil {
		return fmt.Errorf("generate PKCE verifier: %w", err)
	}
	challenge := generateChallenge(verifier)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("start callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		listener.Close()
		return fmt.Errorf("generate state: %w", err)
	}
	state := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(stateBytes)

	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&code_challenge=%s&code_challenge_method=S256&state=%s",
		meta.AuthorizationEndpoint,
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(challenge),
		url.QueryEscape(state),
	)

	fmt.Println("Opening browser for authentication...")
	fmt.Printf("Visit:\n%s\n\n", authURL)
	fmt.Println("Waiting for callback... If on a remote server, paste the redirect URL below:")
	openBrowser(authURL)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errCh <- fmt.Errorf("auth error: %s - %s", errMsg, r.URL.Query().Get("error_description"))
			fmt.Fprintf(w, "<html><body><h2>Authentication failed</h2><p>%s</p><p>You can close this window.</p></body></html>", errMsg)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code received")
			http.Error(w, "No code", http.StatusBadRequest)
			return
		}
		codeCh <- code
		fmt.Fprint(w, "<html><body><h2>Authentication successful!</h2><p>You can close this window and return to the terminal.</p></body></html>")
	})

	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			parsed, err := url.Parse(line)
			if err != nil {
				continue
			}
			code := parsed.Query().Get("code")
			pastedState := parsed.Query().Get("state")
			if code == "" {
				continue
			}
			if pastedState != "" && pastedState != state {
				errCh <- fmt.Errorf("state mismatch in pasted URL")
				return
			}
			codeCh <- code
			return
		}
	}()

	var code string
	select {
	case code = <-codeCh:
		o.log("Received authorization code")
	case err := <-errCh:
		_ = server.Shutdown(ctx)
		return err
	case <-time.After(5 * time.Minute):
		_ = server.Shutdown(ctx)
		return fmt.Errorf("authentication timed out (5 minutes)")
	case <-ctx.Done():
		_ = server.Shutdown(ctx)
		return ctx.Err()
	}

	_ = server.Shutdown(ctx)

	o.log("Exchanging authorization code for tokens...")
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {verifier},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
	}
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", meta.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decode token response: %w", err)
	}

	newTokens := &Tokens{
		AccessToken:   tokenResp.AccessToken,
		RefreshToken:  tokenResp.RefreshToken,
		ExpiresAt:     time.Now().Unix() + tokenResp.ExpiresIn,
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		TokenEndpoint: meta.TokenEndpoint,
	}

	if err := o.Store.Save(newTokens); err != nil {
		return fmt.Errorf("save tokens: %w", err)
	}

	fmt.Println("Login successful!")
	return nil
}

func (o *OAuthConfig) discover(ctx context.Context) (*serverMetadata, error) {
	reqURL := strings.TrimRight(o.ServerURL, "/") + "/.well-known/oauth-authorization-server"
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discovery failed (%d): %s", resp.StatusCode, string(body))
	}

	var meta serverMetadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (o *OAuthConfig) register(ctx context.Context, endpoint string) (*clientRegistration, error) {
	body := map[string]any{
		"redirect_uris": []string{"http://127.0.0.1/callback"},
		"client_name":   "OpenMarkers CLI",
	}
	data, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registration failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var reg clientRegistration
	if err := json.NewDecoder(resp.Body).Decode(&reg); err != nil {
		return nil, err
	}
	return &reg, nil
}

func generateVerifier() (string, error) {
	b := make([]byte, 43)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}

func generateChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h[:])
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		_ = cmd.Start()
	}
}
