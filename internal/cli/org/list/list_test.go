package list

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
)

func stubTokenLoader(tokens auth.TokenCache, err error) auth.TokenLoader {
	return func(config.Config) (auth.TokenCache, error) {
		return tokens, err
	}
}

func cfgWithServer(t *testing.T, handler http.HandlerFunc) config.Config {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := config.Config{}
	cfg.TinaCloud.Endpoint = srv.URL
	return cfg
}

func TestRunList_Unauthenticated(t *testing.T) {
	out, err := runList(config.Config{}, stubTokenLoader(auth.TokenCache{}, auth.ErrUnauthenticated))
	if err != nil {
		t.Fatalf("runList() = %v, want nil error", err)
	}
	if out.LoggedIn {
		t.Errorf("runList().LoggedIn = true, want false")
	}
	if len(out.Organizations) != 0 {
		t.Errorf("runList().Organizations = %v, want empty", out.Organizations)
	}
}

func TestRunList_MultipleOrgs(t *testing.T) {
	orgs := []domain.OrganizationWithRole{
		{Organization: domain.Organization{ID: "id1", Name: "Acme", BillingEnabled: true}, Role: "admin"},
		{Organization: domain.Organization{ID: "id2", Name: "Widgets", BillingEnabled: false}, Role: "member"},
	}
	cfg := cfgWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/v2/organizations" {
			http.NotFound(w, r)
			return
		}
		if got, want := r.Header.Get("Authorization"), "Bearer access-token"; got != want {
			http.Error(w, "bad authorization header", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(orgs)
	})

	out, err := runList(cfg, stubTokenLoader(auth.TokenCache{AccessToken: "access-token"}, nil))
	if err != nil {
		t.Fatalf("runList() = %v, want nil error", err)
	}
	if !out.LoggedIn {
		t.Errorf("runList().LoggedIn = false, want true")
	}
	if len(out.Organizations) != 2 {
		t.Fatalf("runList() returned %d orgs, want 2", len(out.Organizations))
	}
	for i, want := range orgs {
		if got := out.Organizations[i]; got != want {
			t.Errorf("Organizations[%d] = %+v, want %+v", i, got, want)
		}
	}
}

func TestRunList_APIErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    error
	}{
		{name: "unauthorized", statusCode: http.StatusUnauthorized, wantErr: httpclient.ErrUnauthorized},
		{name: "server error", statusCode: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := cfgWithServer(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.statusCode)
			})

			_, err := runList(cfg, stubTokenLoader(auth.TokenCache{AccessToken: "access-token"}, nil))
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("runList() = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err == nil {
				t.Fatal("runList() = nil error, want API error")
			}
		})
	}
}

func TestRunList_TokenLoadError(t *testing.T) {
	_, err := runList(config.Config{}, stubTokenLoader(auth.TokenCache{}, errors.New("disk failure")))
	if err == nil {
		t.Fatal("runList() = nil error, want error for non-auth load failure")
	}
}

func TestRunList_MalformedJSON(t *testing.T) {
	cfg := cfgWithServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{not json"))
	})

	_, err := runList(cfg, stubTokenLoader(auth.TokenCache{AccessToken: "access-token"}, nil))
	if err == nil {
		t.Fatal("runList() = nil error, want unmarshal error")
	}
}
