package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/httpclient"
)

// tokenCachePath redirects the user config dir to a temp dir and returns the
// path where the token cache is expected to live.
func tokenCachePath(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("HOME", dir)            // darwin: $HOME/Library/Application Support
	t.Setenv("XDG_CONFIG_HOME", dir) // linux: $XDG_CONFIG_HOME
	t.Setenv("AppData", dir)         // windows: %AppData%

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("os.UserConfigDir() = %v", err)
	}
	return filepath.Join(configDir, "tinacloud", "cache", "tokens.json")
}

func seedTokenCache(t *testing.T, tokenPath string, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(tokenPath), 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(tokenPath, []byte(contents), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestPeekTokens(t *testing.T) {
	t.Run("returns ErrUnauthenticated when no cache file", func(t *testing.T) {
		tokenCachePath(t)

		_, err := PeekTokens()
		if !errors.Is(err, ErrUnauthenticated) {
			t.Fatalf("PeekTokens() = %v, want ErrUnauthenticated", err)
		}
	})

	t.Run("returns ErrUnauthenticated when tokens are empty", func(t *testing.T) {
		tokenPath := tokenCachePath(t)
		seedTokenCache(t, tokenPath, `{"access_token":"","refresh_token":""}`)

		_, err := PeekTokens()
		if !errors.Is(err, ErrUnauthenticated) {
			t.Fatalf("PeekTokens() = %v, want ErrUnauthenticated", err)
		}
	})

	t.Run("returns expired tokens without refreshing", func(t *testing.T) {
		tokenPath := tokenCachePath(t)
		expired := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
		seedTokenCache(t, tokenPath, `{"access_token":"access-1","refresh_token":"refresh-1","expires_at":"`+expired+`"}`)

		// No server is running anywhere; if PeekTokens attempted a refresh it
		// would need a config endpoint and a network call.
		tokens, err := PeekTokens()
		if err != nil {
			t.Fatalf("PeekTokens() = %v, want nil error", err)
		}
		if tokens.AccessToken != "access-1" {
			t.Errorf("AccessToken = %q, want %q", tokens.AccessToken, "access-1")
		}
		if tokens.RefreshToken != "refresh-1" {
			t.Errorf("RefreshToken = %q, want %q", tokens.RefreshToken, "refresh-1")
		}
	})
}

func TestSaveAndClearTokens(t *testing.T) {
	tokenPath := tokenCachePath(t)

	want := TokenCache{AccessToken: "access-1", RefreshToken: "refresh-1"}
	if err := SaveTokens(want); err != nil {
		t.Fatalf("SaveTokens() = %v, want nil error", err)
	}

	got, err := PeekTokens()
	if err != nil {
		t.Fatalf("PeekTokens() = %v, want nil error", err)
	}
	if got.AccessToken != want.AccessToken || got.RefreshToken != want.RefreshToken {
		t.Errorf("PeekTokens() = %+v, want %+v", got, want)
	}

	if err := ClearTokens(); err != nil {
		t.Fatalf("ClearTokens() = %v, want nil error", err)
	}
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Errorf("os.Stat(%q) = %v, want not-exist error", tokenPath, err)
	}

	// ClearTokens is idempotent: removing an absent file is not an error.
	if err := ClearTokens(); err != nil {
		t.Errorf("ClearTokens() on missing file = %v, want nil error", err)
	}
}

func TestTokensFromCache(t *testing.T) {
	t.Run("returns unexpired tokens without refreshing", func(t *testing.T) {
		tokenPath := tokenCachePath(t)
		future := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		seedTokenCache(t, tokenPath, `{"access_token":"access-1","refresh_token":"refresh-1","expires_at":"`+future+`"}`)

		// Endpoint points nowhere reachable; an unwanted refresh would fail.
		cfg := config.Config{TinaCloud: config.TinaCloudConfig{Endpoint: "http://127.0.0.1:1"}}
		tokens, err := TokensFromCache(cfg)
		if err != nil {
			t.Fatalf("TokensFromCache() = %v, want nil error", err)
		}
		if tokens.AccessToken != "access-1" {
			t.Errorf("AccessToken = %q, want %q", tokens.AccessToken, "access-1")
		}
	})

	t.Run("refreshes expired tokens and saves the result", func(t *testing.T) {
		tokenPath := tokenCachePath(t)
		expired := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
		seedTokenCache(t, tokenPath, `{"access_token":"old-access","refresh_token":"refresh-1","expires_at":"`+expired+`"}`)

		var gotPath, gotGrant, gotRefresh string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotGrant = r.PostFormValue("grant_type")
			gotRefresh = r.PostFormValue("refresh_token")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id_token":"new-id","access_token":"new-access","token_type":"Bearer","expires_in":3600}`))
		}))
		defer server.Close()

		cfg := config.Config{TinaCloud: config.TinaCloudConfig{Endpoint: server.URL}}
		tokens, err := TokensFromCache(cfg)
		if err != nil {
			t.Fatalf("TokensFromCache() = %v, want nil error", err)
		}
		if gotPath != "/oauth/token" {
			t.Errorf("request path = %q, want %q", gotPath, "/oauth/token")
		}
		if gotGrant != "refresh_token" {
			t.Errorf("grant_type = %q, want %q", gotGrant, "refresh_token")
		}
		if gotRefresh != "refresh-1" {
			t.Errorf("refresh_token = %q, want %q", gotRefresh, "refresh-1")
		}
		if tokens.AccessToken != "new-access" {
			t.Errorf("AccessToken = %q, want %q", tokens.AccessToken, "new-access")
		}
		// The refresh token is preserved from the cache, not the response.
		if tokens.RefreshToken != "refresh-1" {
			t.Errorf("RefreshToken = %q, want %q", tokens.RefreshToken, "refresh-1")
		}

		// The refreshed tokens are persisted back to the cache.
		saved, err := PeekTokens()
		if err != nil {
			t.Fatalf("PeekTokens() = %v, want nil error", err)
		}
		if saved.AccessToken != "new-access" {
			t.Errorf("persisted AccessToken = %q, want %q", saved.AccessToken, "new-access")
		}
	})

	t.Run("returns ErrUnauthorized when refresh is rejected", func(t *testing.T) {
		tokenPath := tokenCachePath(t)
		expired := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
		seedTokenCache(t, tokenPath, `{"access_token":"old-access","refresh_token":"refresh-1","expires_at":"`+expired+`"}`)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		cfg := config.Config{TinaCloud: config.TinaCloudConfig{Endpoint: server.URL}}
		if _, err := TokensFromCache(cfg); !errors.Is(err, httpclient.ErrUnauthorized) {
			t.Fatalf("TokensFromCache() = %v, want ErrUnauthorized", err)
		}
	})
}

func TestRevokeToken(t *testing.T) {
	t.Run("posts refresh token to revoke endpoint", func(t *testing.T) {
		var gotPath, gotToken string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotToken = r.PostFormValue("token")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := config.Config{TinaCloud: config.TinaCloudConfig{Endpoint: server.URL}}

		if err := RevokeToken(cfg, "refresh-1"); err != nil {
			t.Fatalf("RevokeToken() = %v, want nil error", err)
		}
		if gotPath != "/oauth/revoke" {
			t.Errorf("request path = %q, want %q", gotPath, "/oauth/revoke")
		}
		if gotToken != "refresh-1" {
			t.Errorf("token form value = %q, want %q", gotToken, "refresh-1")
		}
	})

	t.Run("returns error on non-2xx response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		cfg := config.Config{TinaCloud: config.TinaCloudConfig{Endpoint: server.URL}}

		if err := RevokeToken(cfg, "refresh-1"); err == nil {
			t.Fatal("RevokeToken() = nil error, want error")
		}
	})

	t.Run("returns error when server unreachable", func(t *testing.T) {
		cfg := config.Config{TinaCloud: config.TinaCloudConfig{Endpoint: "http://127.0.0.1:1"}}

		if err := RevokeToken(cfg, "refresh-1"); err == nil {
			t.Fatal("RevokeToken() = nil error, want error")
		}
	})
}
