package config

import (
	"context"
	"os"
	"testing"
)

// unset clears an environment variable for the duration of the test, restoring
// its original value on cleanup via t.Setenv's restore hook.
func unset(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	_ = os.Unsetenv(key)
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want Config
	}{
		{
			name: "defaults",
			want: Config{
				Core:      CoreConfig{Dev: true},
				TinaCloud: TinaCloudConfig{Endpoint: "https://identity-v2.tinajs.io", FrontEndEndpoint: "https://app.tina.io"},
			},
		},
		{
			name: "env override",
			env: map[string]string{
				"DEV":                "false",
				"POSTHOG_API_KEY":    "ph-key",
				"POSTHOG_ENDPOINT":   "https://ph.example",
				"TINACLOUD_ENDPOINT": "https://tc.example",
			},
			want: Config{
				Core:      CoreConfig{Dev: false},
				TinaCloud: TinaCloudConfig{Endpoint: "https://tc.example", FrontEndEndpoint: "https://app.tina.io"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{"DEV", "POSTHOG_API_KEY", "POSTHOG_ENDPOINT", "TINACLOUD_ENDPOINT", "TINACLOUD_FRONT_END_ENDPOINT"} {
				unset(t, key)
			}
			for key, val := range tc.env {
				t.Setenv(key, val)
			}

			got, err := LoadConfig(context.Background())
			if err != nil {
				t.Fatalf("LoadConfig() = %v, want nil error", err)
			}
			if got != tc.want {
				t.Errorf("LoadConfig() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
