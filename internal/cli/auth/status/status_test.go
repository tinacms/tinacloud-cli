package status

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v4"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/httpclient"
)

func makeJWT(t *testing.T, userID string) string {
	t.Helper()

	claims := &auth.DecodedAccessToken{
		TinaCloudUserID:  userID,
		RegisteredClaims: jwt.RegisteredClaims{},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("makeJWT: %v", err)
	}
	return signed
}

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

func TestRunStatus_Unauthenticated(t *testing.T) {
	out, err := runStatus(config.Config{}, stubTokenLoader(auth.TokenCache{}, auth.ErrUnauthenticated))
	if err != nil {
		t.Fatalf("runStatus() = %v, want nil error", err)
	}
	if out.LoggedIn {
		t.Errorf("runStatus().LoggedIn = true, want false")
	}
}

func TestRunStatus_Success(t *testing.T) {
	accessToken := makeJWT(t, "user-123")
	cfg := cfgWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/v2/users/user-123" {
			http.NotFound(w, r)
			return
		}
		if got, want := r.Header.Get("Authorization"), "Bearer "+accessToken; got != want {
			http.Error(w, "bad authorization header", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"fullName": "Alice Smith",
			"email":    "alice@example.com",
		})
	})

	tokens := auth.TokenCache{AccessToken: accessToken}
	out, err := runStatus(cfg, stubTokenLoader(tokens, nil))
	if err != nil {
		t.Fatalf("runStatus() = %v, want nil error", err)
	}
	if !out.LoggedIn {
		t.Errorf("runStatus().LoggedIn = false, want true")
	}
	if out.FullName != "Alice Smith" {
		t.Errorf("runStatus().FullName = %q, want %q", out.FullName, "Alice Smith")
	}
	if out.Email != "alice@example.com" {
		t.Errorf("runStatus().Email = %q, want %q", out.Email, "alice@example.com")
	}
}

func TestRunStatus_APIErrors(t *testing.T) {
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

			tokens := auth.TokenCache{AccessToken: makeJWT(t, "user-123")}
			_, err := runStatus(cfg, stubTokenLoader(tokens, nil))
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("runStatus() = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err == nil {
				t.Fatal("runStatus() = nil error, want API error")
			}
		})
	}
}

func TestRunStatus_NoUserID(t *testing.T) {
	tokens := auth.TokenCache{AccessToken: makeJWT(t, "")}
	_, err := runStatus(config.Config{}, stubTokenLoader(tokens, nil))
	if err == nil {
		t.Fatal("runStatus() = nil error, want error for missing user ID")
	}
}

func TestRunStatus_TokenLoadError(t *testing.T) {
	_, err := runStatus(config.Config{}, stubTokenLoader(auth.TokenCache{}, errors.New("disk failure")))
	if err == nil {
		t.Fatal("runStatus() = nil error, want error for non-auth load failure")
	}
}

func TestRunStatus_DecodeError(t *testing.T) {
	tokens := auth.TokenCache{AccessToken: "not-a-jwt"}
	_, err := runStatus(config.Config{}, stubTokenLoader(tokens, nil))
	if err == nil {
		t.Fatal("runStatus() = nil error, want decode error for malformed token")
	}
}

func TestRunStatus_MalformedJSON(t *testing.T) {
	cfg := cfgWithServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{not json"))
	})

	tokens := auth.TokenCache{AccessToken: makeJWT(t, "user-123")}
	_, err := runStatus(cfg, stubTokenLoader(tokens, nil))
	if err == nil {
		t.Fatal("runStatus() = nil error, want unmarshal error")
	}
}
