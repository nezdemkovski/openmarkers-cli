package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestAuthorizationServerMetadataURLsForPathIssuer(t *testing.T) {
	got := authorizationServerMetadataURLs("https://auth.nezdemkovski.cloud/api/openmarkers")
	want := []string{
		"https://auth.nezdemkovski.cloud/api/openmarkers/.well-known/openid-configuration",
		"https://auth.nezdemkovski.cloud/api/openmarkers/.well-known/oauth-authorization-server",
		"https://auth.nezdemkovski.cloud/.well-known/oauth-authorization-server/api/openmarkers",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("authorizationServerMetadataURLs() = %#v, want %#v", got, want)
	}
}

func TestAuthorizationServerMetadataURLsRejectsInvalidIssuer(t *testing.T) {
	if got := authorizationServerMetadataURLs("openmarkers"); got != nil {
		t.Fatalf("authorizationServerMetadataURLs() = %#v, want nil", got)
	}
}

func TestDiscoverUsesProtectedResourceAndAuthorizationServerMetadata(t *testing.T) {
	var authServer *httptest.Server
	authServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/openmarkers/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":                 authServer.URL + "/api/openmarkers",
			"authorization_endpoint": authServer.URL + "/api/openmarkers/auth/oauth2/authorize",
			"token_endpoint":         authServer.URL + "/api/openmarkers/auth/oauth2/token",
			"registration_endpoint":  authServer.URL + "/api/openmarkers/auth/oauth2/register",
		})
	}))
	defer authServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/oauth-protected-resource" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"resource":              apiServerResource,
			"authorization_servers": []string{authServer.URL + "/api/openmarkers"},
		})
	}))
	defer apiServer.Close()

	config := &OAuthConfig{ServerURL: apiServer.URL}
	meta, err := config.discover(context.Background())
	if err != nil {
		t.Fatalf("discover() error = %v", err)
	}

	if meta.Resource != apiServerResource {
		t.Fatalf("Resource = %q, want %q", meta.Resource, apiServerResource)
	}
	if meta.AuthorizationEndpoint != authServer.URL+"/api/openmarkers/auth/oauth2/authorize" {
		t.Fatalf("AuthorizationEndpoint = %q", meta.AuthorizationEndpoint)
	}
}

func TestRegisterUsesPublicPKCEClientWithExactRedirectURI(t *testing.T) {
	const redirectURI = "http://127.0.0.1:49152/callback"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode registration body: %v", err)
		}

		redirects, ok := body["redirect_uris"].([]any)
		if !ok || len(redirects) != 1 || redirects[0] != redirectURI {
			t.Fatalf("redirect_uris = %#v", body["redirect_uris"])
		}
		if body["token_endpoint_auth_method"] != "none" {
			t.Fatalf("token_endpoint_auth_method = %#v", body["token_endpoint_auth_method"])
		}
		if body["scope"] != "openid profile email offline_access" {
			t.Fatalf("scope = %#v", body["scope"])
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"client_id": "cli-test"})
	}))
	defer server.Close()

	config := &OAuthConfig{}
	reg, err := config.register(context.Background(), server.URL, redirectURI)
	if err != nil {
		t.Fatalf("register() error = %v", err)
	}
	if reg.ClientID != "cli-test" {
		t.Fatalf("ClientID = %q", reg.ClientID)
	}
}

const apiServerResource = "https://openmarkers.app/mcp"
