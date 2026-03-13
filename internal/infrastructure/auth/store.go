package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/openmarkers/openmarkers-cli/internal/infrastructure/secrets"
)

type Tokens struct {
	AccessToken   string `json:"access_token,omitempty"`
	RefreshToken  string `json:"refresh_token,omitempty"`
	ExpiresAt     int64  `json:"expires_at"`
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret,omitempty"`
	TokenEndpoint string `json:"token_endpoint"`
}

type authMeta struct {
	ExpiresAt     int64  `json:"expires_at"`
	ClientID      string `json:"client_id"`
	TokenEndpoint string `json:"token_endpoint"`
}

type Store struct {
	path   string
	mu     sync.Mutex
	cached *Tokens
}

func NewStore(configDir string) *Store {
	return &Store{
		path: filepath.Join(configDir, "auth.json"),
	}
}

func (s *Store) Load() *Tokens {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cached != nil {
		t := *s.cached
		return &t
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil
	}

	var meta authMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil
	}

	accessToken, _, _ := secrets.Get(secrets.KeyAccessToken)
	refreshToken, _, _ := secrets.Get(secrets.KeyRefreshToken)
	clientSecret, _, _ := secrets.Get(secrets.KeyClientSecret)

	tokens := &Tokens{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ExpiresAt:     meta.ExpiresAt,
		ClientID:      meta.ClientID,
		ClientSecret:  clientSecret,
		TokenEndpoint: meta.TokenEndpoint,
	}

	s.cached = tokens
	t := *s.cached
	return &t
}

func (s *Store) Save(tokens *Tokens) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	warnedInsecure := false
	warnInsecure := func(insecure bool) {
		if insecure && !warnedInsecure {
			fmt.Fprintln(os.Stderr, "Warning: storing tokens in plaintext file (keyring unavailable)")
			warnedInsecure = true
		}
	}

	if tokens.AccessToken != "" {
		if _, insecure, err := secrets.Set(secrets.KeyAccessToken, tokens.AccessToken); err != nil {
			return fmt.Errorf("store access token: %w", err)
		} else {
			warnInsecure(insecure)
		}
	}
	if tokens.RefreshToken != "" {
		if _, insecure, err := secrets.Set(secrets.KeyRefreshToken, tokens.RefreshToken); err != nil {
			return fmt.Errorf("store refresh token: %w", err)
		} else {
			warnInsecure(insecure)
		}
	}
	if tokens.ClientSecret != "" {
		if _, insecure, err := secrets.Set(secrets.KeyClientSecret, tokens.ClientSecret); err != nil {
			return fmt.Errorf("store client secret: %w", err)
		} else {
			warnInsecure(insecure)
		}
	}

	meta := authMeta{
		ExpiresAt:     tokens.ExpiresAt,
		ClientID:      tokens.ClientID,
		TokenEndpoint: tokens.TokenEndpoint,
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	s.cached = tokens
	return os.WriteFile(s.path, data, 0o600)
}

func (s *Store) Delete() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cached = nil

	var errs []error
	if err := secrets.DeleteAll(); err != nil {
		errs = append(errs, err)
	}

	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (s *Store) HasTokens() bool {
	tokens := s.Load()
	return tokens != nil && tokens.AccessToken != ""
}
