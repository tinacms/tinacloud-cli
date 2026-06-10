package logout

import (
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
		name     string
		seedFile bool
	}{
		{name: "removes existing token file", seedFile: true},
		{name: "succeeds when no token file", seedFile: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokenPath := tokenCachePath(t)
			if tc.seedFile {
				if err := os.MkdirAll(filepath.Dir(tokenPath), 0700); err != nil {
					t.Fatalf("MkdirAll: %v", err)
				}
				if err := os.WriteFile(tokenPath, []byte(`{"access_token":"x"}`), 0600); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			}

			if err := Logout(nil, config.Config{}); err != nil {
				t.Fatalf("Logout() = %v, want nil error", err)
			}

			if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
				t.Errorf("token file present after Logout, stat err = %v", err)
			}
		})
	}
}
