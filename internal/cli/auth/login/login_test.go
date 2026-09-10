package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
	"tina.io/tinacloud-cli/internal/posthog"
	"tina.io/tinacloud-cli/internal/utils"
)

// captureTracker records emitted events and identify calls for assertions.
type captureTracker struct {
	events     []posthog.Capture
	identified []string
}

func (c *captureTracker) track(event posthog.Capture) { c.events = append(c.events, event) }

func (c *captureTracker) identify(userID string) { c.identified = append(c.identified, userID) }

func (c *captureTracker) count(event string) int {
	n := 0
	for _, e := range c.events {
		if e.Event == event {
			n++
		}
	}
	return n
}

// fakeAccessToken is a JWT carrying fakeUserID; the signature is unverified, so the secret is irrelevant.
const fakeUserID = "user-abc"

var fakeAccessToken = func() string {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"external_id": fakeUserID})
	signed, _ := tok.SignedString([]byte("test-secret"))
	return signed
}()

func TestGenerateChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	got := generateChallenge(verifier)
	if got != want {
		t.Errorf("generateChallenge(%q) = %q, want %q", verifier, got, want)
	}
}

func TestGenerateVerifier_Format(t *testing.T) {
	verifier, err := generateVerifier()
	if err != nil {
		t.Fatalf("generateVerifier() = %v, want nil", err)
	}

	if verifier == "" {
		t.Fatal(`generateVerifier() = "", want non-empty string`)
	}
	if len(verifier) != 43 {
		t.Errorf("len(generateVerifier()) = %d, want 43", len(verifier))
	}

	// Raw URL encoding must not contain '+', '/', or '='.
	for _, ch := range []string{"+", "/", "="} {
		if strings.Contains(verifier, ch) {
			t.Errorf("generateVerifier() = %q contains disallowed character %q", verifier, ch)
		}
	}
}

func TestCallbackListener(t *testing.T) {
	ln, port, err := utils.CallbackListener()
	if err != nil {
		t.Fatalf("callbackListener() = %v, want nil error", err)
	}
	defer ln.Close()
	if port < 9999 || port > 10011 {
		t.Errorf("callbackListener() port = %d, want a port in [9999, 10011]", port)
	}
}

// Login should fail before opening a browser when the authorize request to
// TinaCloud cannot be made. Pointing the endpoint at a closed port exercises
// that error path without any browser side effects.
func TestLogin_AuthorizeRequestFails(t *testing.T) {
	cfg := config.Config{}
	cfg.TinaCloud.Endpoint = "http://127.0.0.1:1"

	err := runLogin(cfg, nil, posthog.NoopTracker, posthog.NoopIdentifier)
	if err == nil {
		t.Fatal("Login() = nil error, want error when authorize request fails")
	}
	if !strings.Contains(err.Error(), "failed to authorize with tinacloud") {
		t.Errorf("Login() = %v, want authorize failure", err)
	}
}

// validCallback is a well-formed OAuth redirect carrying an auth code.
const validCallback = "GET /callback?code=test-code HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"

// driveLogin runs runLogin against a fake TinaCloud server with a fake browser
// opener. tokenHandler serves /oauth/token; callbackReq is the raw HTTP request
// delivered to the local callback server in place of a real browser redirect.
func driveLogin(t *testing.T, tokenHandler http.HandlerFunc, callbackReq string) (*captureTracker, error) {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("HOME", dir)            // darwin token cache location
	t.Setenv("XDG_CONFIG_HOME", dir) // linux token cache location
	t.Setenv("AppData", dir)         // windows token cache location

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/authorize":
			// No Location header: runLogin falls back to the authorize URL,
			// which carries the callback port we drive below.
			w.WriteHeader(http.StatusOK)
		case "/oauth/token":
			tokenHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	// Fake browser: parse the callback port from the redirect URL and deliver
	// the callback request, simulating the user completing SSO.
	fakeOpener := func(rawURL string) error {
		u, err := url.Parse(rawURL)
		if err != nil {
			return err
		}
		cb, err := url.Parse(u.Query().Get("redirect_uri"))
		if err != nil {
			return err
		}
		go deliverCallback(cb.Host, callbackReq)
		return nil
	}

	cfg := config.Config{}
	cfg.TinaCloud.Endpoint = srv.URL
	tracker := &captureTracker{}
	return tracker, runLogin(cfg, fakeOpener, tracker.track, tracker.identify)
}

// deliverCallback connects to the local callback server and sends request,
// retrying until the server is listening.
func deliverCallback(host, request string) {
	var conn net.Conn
	for i := 0; i < 100; i++ {
		c, err := net.Dial("tcp", host)
		if err == nil {
			conn = c
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if conn == nil {
		return
	}
	defer conn.Close()

	fmt.Fprint(conn, request)
	_, _ = io.ReadAll(conn)
}

func fakeTokenHandler(w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(domain.AuthResponse{
		IDToken:      "id-token",
		AccessToken:  fakeAccessToken,
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	})
}

func TestLogin_Success(t *testing.T) {
	tracker, err := driveLogin(t, fakeTokenHandler, validCallback)
	if err != nil {
		t.Fatalf("Login() = %v, want nil error", err)
	}

	if n := tracker.count(posthog.UserSignInSuccessEvent); n != 1 {
		t.Errorf("UserSignInSuccess emitted %d times, want 1", n)
	}
	if want := []string{fakeUserID}; len(tracker.identified) != 1 || tracker.identified[0] != fakeUserID {
		t.Errorf("identified = %v, want %v", tracker.identified, want)
	}

	tokens, err := auth.TokensFromCache(config.Config{})
	if err != nil {
		t.Fatalf("TokensFromCache() = %v, want nil error", err)
	}
	if tokens.AccessToken != fakeAccessToken {
		t.Errorf("saved AccessToken = %q, want %q", tokens.AccessToken, fakeAccessToken)
	}
	if tokens.RefreshToken != "refresh-token" {
		t.Errorf("saved RefreshToken = %q, want %q", tokens.RefreshToken, "refresh-token")
	}
}

func TestLogin_Errors(t *testing.T) {
	tests := []struct {
		name            string
		tokenHandler    http.HandlerFunc
		callbackReq     string
		wantErr         error
		wantErrContains string
	}{
		{
			name:            "missing code in callback",
			tokenHandler:    fakeTokenHandler,
			callbackReq:     "GET /callback HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n",
			wantErrContains: "code is missing",
		},
		{
			name:            "malformed callback request",
			tokenHandler:    fakeTokenHandler,
			callbackReq:     "this is not a valid http request\r\n\r\n",
			wantErrContains: "failed to read auth callback",
		},
		{
			name:         "token endpoint unauthorized",
			tokenHandler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusUnauthorized) },
			callbackReq:  validCallback,
			wantErr:      httpclient.ErrUnauthorized,
		},
		{
			name:            "token endpoint server error",
			tokenHandler:    func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) },
			callbackReq:     validCallback,
			wantErrContains: "received response status of '500'",
		},
		{
			name:            "token endpoint invalid json",
			tokenHandler:    func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{not json")) },
			callbackReq:     validCallback,
			wantErrContains: "could not parse tokens",
		},
		{
			name:            "token endpoint empty tokens",
			tokenHandler:    func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{}")) },
			callbackReq:     validCallback,
			wantErrContains: "received empty auth tokens",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tracker, err := driveLogin(t, tc.tokenHandler, tc.callbackReq)
			if err == nil {
				t.Fatalf("Login() = nil error, want error")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Errorf("Login() = %v, want errors.Is %v", err, tc.wantErr)
			}
			if tc.wantErrContains != "" && !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Errorf("Login() = %v, want error containing %q", err, tc.wantErrContains)
			}
			if n := tracker.count(posthog.UserSignInSuccessEvent); n != 0 {
				t.Errorf("UserSignInSuccess emitted %d times on a failed login, want 0", n)
			}
			if len(tracker.identified) != 0 {
				t.Errorf("identified %v on a failed login, want none", tracker.identified)
			}
		})
	}
}
