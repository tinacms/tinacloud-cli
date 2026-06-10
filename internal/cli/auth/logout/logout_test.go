package logout

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"tina.io/tinacloud-cli/internal/config"
)

// tokenCachePath redirects the user config dir to a temp dir and returns the
// path where ClearTokens expects the token cache to live.
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

func TestLogout(t *testing.T) {
	tests := []struct {
		name         string
		seedFile     bool
		revokeStatus int
		wantRevoked  bool
	}{
		{
			name:         "revokes token upstream and removes token file",
			seedFile:     true,
			revokeStatus: http.StatusOK,
			wantRevoked:  true,
		},
		{
			name:         "still logs out when upstream revocation fails",
			seedFile:     true,
			revokeStatus: http.StatusInternalServerError,
			wantRevoked:  true,
		},
		{
			name:     "succeeds without revoking when no token file",
			seedFile: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokenPath := tokenCachePath(t)
			if tc.seedFile {
				if err := os.MkdirAll(filepath.Dir(tokenPath), 0700); err != nil {
					t.Fatalf("MkdirAll: %v", err)
				}
				if err := os.WriteFile(tokenPath, []byte(`{"access_token":"access-1","refresh_token":"refresh-1"}`), 0600); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			}

			gotToken := ""
			revokeCalls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				revokeCalls++
				gotToken = r.PostFormValue("token")
				w.WriteHeader(tc.revokeStatus)
			}))
			defer server.Close()

			cfg := config.Config{TinaCloud: config.TinaCloudConfig{Endpoint: server.URL}}

			if err := Logout(nil, cfg); err != nil {
				t.Fatalf("Logout() = %v, want nil error", err)
			}

			if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
				t.Errorf("token file present after Logout, stat err = %v", err)
			}

			if tc.wantRevoked {
				if revokeCalls != 1 {
					t.Errorf("revoke endpoint called %d times, want 1", revokeCalls)
				}
				if gotToken != "refresh-1" {
					t.Errorf("revoked token = %q, want %q", gotToken, "refresh-1")
				}
			} else if revokeCalls != 0 {
				t.Errorf("revoke endpoint called %d times, want 0", revokeCalls)
			}
		})
	}
}
