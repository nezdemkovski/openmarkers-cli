package auth

import (
	"encoding/json"
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
		return s.cached
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
	return s.cached
}

func (s *Store) Save(tokens *Tokens) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tokens.AccessToken != "" {
		if _, _, err := secrets.Set(secrets.KeyAccessToken, tokens.AccessToken); err != nil {
			return fmt.Errorf("store access token: %w", err)
		}
	}
	if tokens.RefreshToken != "" {
		if _, _, err := secrets.Set(secrets.KeyRefreshToken, tokens.RefreshToken); err != nil {
			return fmt.Errorf("store refresh token: %w", err)
		}
	}
	if tokens.ClientSecret != "" {
		if _, _, err := secrets.Set(secrets.KeyClientSecret, tokens.ClientSecret); err != nil {
			return fmt.Errorf("store client secret: %w", err)
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

	secrets.DeleteAll()

	return os.Remove(s.path)
}

func (s *Store) HasTokens() bool {
	tokens := s.Load()
	return tokens != nil && tokens.AccessToken != ""
}
