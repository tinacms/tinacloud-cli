package leaderboard

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
)

func TestCommandLeaderboard(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("AppData", dir)
	if err := auth.SaveTokens(auth.TokenCache{
		AccessToken: "access-token", RefreshToken: "refresh-token", ExpiresAt: time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatal(err)
	}

	requests := make(chan *http.Request, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r
		fmt.Fprint(w, `{"entries":[{"rank":1,"name":"Active User","count":3},{"rank":2,"name":"Inactive User","count":0}]}`)
	}))
	defer server.Close()
	cfg := config.Config{}
	cfg.TinaCloud.ContentEndpoint = server.URL

	for _, tc := range []struct {
		name         string
		flags        []string
		wantInactive bool
	}{
		{name: "default"},
		{name: "all", flags: []string{"--all"}, wantInactive: true},
		{name: "json", flags: []string{"--json"}},
		{name: "json all", flags: []string{"--json", "--all"}, wantInactive: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "output")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			stdout := os.Stdout
			os.Stdout = output
			defer func() { os.Stdout = stdout }()

			args := append([]string{"--org-id", "org-123", "--from", "2025-02-03", "--to", "2025-02-20"}, tc.flags...)
			if err := Command(cfg).Dispatch(args, []string{"org", "leaderboard"}); err != nil {
				t.Fatalf("Dispatch() = %v", err)
			}
			r := <-requests
			if r.Method != http.MethodGet || r.URL.Path != "/leaderboard/org/org-123" || r.URL.Query().Get("from") != "2025-02-03" || r.URL.Query().Get("to") != "2025-02-20" {
				t.Errorf("request = %s %s, want leaderboard for 2025-02-03 through 2025-02-20", r.Method, r.URL.String())
			}
			if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
				t.Errorf("Authorization = %q, want Bearer access-token", got)
			}
			data, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), "Active User") || strings.Contains(string(data), "Inactive User") != tc.wantInactive {
				t.Errorf("output = %s, want active user and inactive user included = %t", data, tc.wantInactive)
			}
		})
	}
}
