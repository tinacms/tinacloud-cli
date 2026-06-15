package configure

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	ph "github.com/posthog/posthog-go"
	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/posthog"
)

// stubLoader returns a TokenLoader that yields the given tokens and error.
func stubLoader(tokens auth.TokenCache, err error) auth.TokenLoader {
	return func(config.Config) (auth.TokenCache, error) { return tokens, err }
}

func TestRepoSlugFromURL(t *testing.T) {
	tests := []struct {
		name    string
		repoURL string
		want    string
	}{
		{
			name:    "ssh with .git suffix",
			repoURL: "git@github.com:tinacms/tinacloud-cli.git",
			want:    "tinacms/tinacloud-cli",
		},
		{
			name:    "https with .git suffix",
			repoURL: "https://github.com/tinacms/tinacloud-cli.git",
			want:    "tinacms/tinacloud-cli",
		},
		{
			name:    "https without .git suffix",
			repoURL: "https://github.com/tinacms/tinacloud-cli",
			want:    "tinacms/tinacloud-cli",
		},
		{
			name:    "fewer than two path segments",
			repoURL: "tinacloud-cli",
			want:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := repoSlugFromURL(tc.repoURL); got != tc.want {
				t.Errorf("repoSlugFromURL(%q) = %q, want %q", tc.repoURL, got, tc.want)
			}
		})
	}
}

func TestRepoSlugFromGitConfig(t *testing.T) {
	cwd := t.TempDir()
	if err := os.Mkdir(path.Join(cwd, ".git"), 0700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	gitConfig := "[remote \"origin\"]\n\turl = git@github.com:tinacms/tinacloud-cli.git\n\tfetch = +refs/heads/*\n"
	if err := os.WriteFile(path.Join(cwd, ".git", "config"), []byte(gitConfig), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := repoSlugFromGitConfig(cwd)
	if err != nil {
		t.Fatalf("repoSlugFromGitConfig() = %v, want nil error", err)
	}
	if want := "tinacms/tinacloud-cli"; got != want {
		t.Errorf("repoSlugFromGitConfig() = %q, want %q", got, want)
	}
}

func TestRunConfigureFlags(t *testing.T) {
	t.Run("writes config when both ids are passed", func(t *testing.T) {
		cwd := t.TempDir()
		if err := os.Mkdir(path.Join(cwd, "tina"), 0700); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}

		// loadTokens must not be reached on the flag path.
		loader := func(config.Config) (auth.TokenCache, error) {
			t.Error("loadTokens should not be called when both ids are passed")
			return auth.TokenCache{}, errors.New("unexpected")
		}

		saved, err := runConfigure(config.Config{}, cwd, "org-1", "project-1", "tina", loader, posthog.NoopTracker)
		if err != nil {
			t.Fatalf("runConfigure() = %v, want nil error", err)
		}
		if !saved {
			t.Error("runConfigure() saved = false, want true")
		}

		got, err := domain.TinaCloudConfigFromDisk(cwd, "tina")
		if err != nil {
			t.Fatalf("TinaCloudConfigFromDisk() = %v, want nil error", err)
		}
		if want := (domain.TinaCloudConfig{OrgID: "org-1", ProjectID: "project-1"}); got != want {
			t.Errorf("written config = %+v, want %+v", got, want)
		}
	})

	t.Run("errors when only one id is passed", func(t *testing.T) {
		saved, err := runConfigure(config.Config{}, t.TempDir(), "org-1", "", "tina", stubLoader(auth.TokenCache{}, nil), posthog.NoopTracker)
		if err == nil {
			t.Fatal("runConfigure() = nil error, want error")
		}
		if saved {
			t.Error("runConfigure() saved = true, want false")
		}
	})
}

func TestRunConfigureUnauthenticated(t *testing.T) {
	saved, err := runConfigure(config.Config{}, t.TempDir(), "", "", "tina", stubLoader(auth.TokenCache{}, auth.ErrUnauthenticated), posthog.NoopTracker)
	if !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("runConfigure() = %v, want ErrUnauthenticated", err)
	}
	if saved {
		t.Error("runConfigure() saved = true, want false")
	}
}

// TestRunConfigureTracking verifies the event is emitted only after a successful save.
func TestRunConfigureTracking(t *testing.T) {
	t.Run("emits user-configured after a successful save", func(t *testing.T) {
		cwd := t.TempDir()
		if err := os.Mkdir(path.Join(cwd, "tina"), 0700); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}

		var got []ph.Capture
		track := func(c ph.Capture) { got = append(got, c) }

		if _, err := runConfigure(config.Config{}, cwd, "org-1", "project-1", "tina", nil, track); err != nil {
			t.Fatalf("runConfigure() = %v, want nil error", err)
		}

		if len(got) != 1 {
			t.Fatalf("tracked %d events, want 1", len(got))
		}
		if got[0].Event != posthog.UserConfiguredEvent {
			t.Errorf("event = %q, want %q", got[0].Event, posthog.UserConfiguredEvent)
		}
		if method := got[0].Properties["configuration_method"]; method != posthog.ConfigMethodManualSelect {
			t.Errorf("configuration_method = %v, want %q", method, posthog.ConfigMethodManualSelect)
		}
	})

	t.Run("does not emit when the save fails", func(t *testing.T) {
		// A missing tina dir makes the save fail, so nothing should be tracked.
		var got []ph.Capture
		track := func(c ph.Capture) { got = append(got, c) }

		if _, err := runConfigure(config.Config{}, t.TempDir(), "org-1", "project-1", "tina", nil, track); err == nil {
			t.Fatal("runConfigure() = nil error, want error")
		}

		if len(got) != 0 {
			t.Errorf("tracked %d events, want 0", len(got))
		}
	})
}

func TestRunConfigureBackend(t *testing.T) {
	t.Run("returns error when organizations request fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := config.Config{TinaCloud: config.TinaCloudConfig{Endpoint: server.URL}}
		if _, err := runConfigure(cfg, t.TempDir(), "", "", "tina", stubLoader(auth.TokenCache{AccessToken: "tok"}, nil), posthog.NoopTracker); err == nil {
			t.Fatal("runConfigure() = nil error, want error")
		}
	})

	t.Run("errors when the account has no projects", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/v2/organizations":
				_, _ = w.Write([]byte(`[{"organization":{"id":"org-1","name":"Acme"}}]`))
			case strings.HasSuffix(r.URL.Path, "/apps"):
				_, _ = w.Write([]byte(`[]`))
			default:
				t.Errorf("unexpected request path %q", r.URL.Path)
			}
		}))
		defer server.Close()

		cfg := config.Config{TinaCloud: config.TinaCloudConfig{Endpoint: server.URL}}
		_, err := runConfigure(cfg, t.TempDir(), "", "", "tina", stubLoader(auth.TokenCache{AccessToken: "tok"}, nil), posthog.NoopTracker)
		if err == nil || !strings.Contains(err.Error(), "account has no projects") {
			t.Fatalf("runConfigure() = %v, want \"account has no projects\" error", err)
		}
	})
}
