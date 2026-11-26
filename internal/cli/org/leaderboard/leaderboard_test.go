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

func TestCommandRequiresBothDates(t *testing.T) {
	for _, flags := range [][]string{{"--from", "2025-01-01"}, {"--to", "2025-01-31"}} {
		err := Command(config.Config{}).Dispatch(flags, nil)
		if err == nil || err.Error() != "--from and --to must be used together" {
			t.Errorf("Dispatch(%v) = %v, want paired date error", flags, err)
		}
	}
}

func TestPeriodDates(t *testing.T) {
	now := time.Date(2024, time.March, 31, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		period, from, to string
	}{
		{period: "this-month", from: "2024-03-01", to: "2024-03-31"},
		{period: "last-month", from: "2024-02-01", to: "2024-02-29"},
		{period: "3-months", from: "2024-01-01", to: "2024-03-31"},
		{period: "12-months", from: "2023-04-01", to: "2024-03-31"},
	} {
		t.Run(tc.period, func(t *testing.T) {
			from, to, err := periodDates(tc.period, now)
			if err != nil || from != tc.from || to != tc.to {
				t.Errorf("periodDates(%q) = %q, %q, %v; want %q, %q, nil", tc.period, from, to, err, tc.from, tc.to)
			}
		})
	}
	if _, _, err := periodDates("invalid", now); err == nil {
		t.Error("periodDates(invalid) succeeded, want error")
	}
}

func TestCommandPeriod(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("AppData", dir)
	if err := auth.SaveTokens(auth.TokenCache{AccessToken: "access-token", RefreshToken: "refresh-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	requests := make(chan *http.Request, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r
		fmt.Fprint(w, `{"entries":[]}`)
	}))
	defer server.Close()
	cfg := config.Config{}
	cfg.TinaCloud.ContentEndpoint = server.URL

	if err := Command(cfg).Dispatch([]string{"--org-id", "org-123", "--period", "last-month", "--json"}, nil); err != nil {
		t.Fatal(err)
	}
	r := <-requests
	from, to, _ := periodDates("last-month", time.Now())
	if r.URL.Query().Get("from") != from || r.URL.Query().Get("to") != to {
		t.Errorf("request = %s, want from=%s and to=%s", r.URL, from, to)
	}
	for _, flags := range [][]string{{"--period", "invalid"}, {"--period", "this-month", "--from", "2025-01-01"}, {"--period", "this-month", "--to", "2025-01-01"}} {
		if err := Command(cfg).Dispatch(flags, nil); err == nil {
			t.Errorf("Dispatch(%v) succeeded, want error", flags)
		}
	}
}
