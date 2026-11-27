package leaderboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
	requestapi "tina.io/tinacloud-cli/internal/requests"
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

			args := append([]string{"--org-id", "11111111-1111-4111-8111-111111111111", "--from", "2025-02-03", "--to", "2025-02-20"}, tc.flags...)
			if err := Command(cfg, requestapi.Organization).Dispatch(args, []string{"org", "leaderboard"}); err != nil {
				t.Fatalf("Dispatch() = %v", err)
			}
			r := <-requests
			if r.Method != http.MethodGet || r.URL.Path != "/leaderboard/org/11111111-1111-4111-8111-111111111111" || r.URL.Query().Get("from") != "2025-02-03" || r.URL.Query().Get("to") != "2025-02-20" {
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
		err := Command(config.Config{}, requestapi.Organization).Dispatch(flags, nil)
		if err == nil || err.Error() != "--from and --to must be used together" {
			t.Errorf("Dispatch(%v) = %v, want paired date error", flags, err)
		}
	}
}

func TestCommandRejectsInvalidDatesBeforeRequest(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("AppData", dir)
	if err := auth.SaveTokens(auth.TokenCache{AccessToken: "access-token", RefreshToken: "refresh-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("unexpected HTTP request")
	}))
	defer server.Close()
	cfg := config.Config{}
	cfg.TinaCloud.ContentEndpoint = server.URL

	for _, tc := range []struct {
		name, from, to, want string
	}{
		{name: "invalid from", from: "2026-02-30", to: "2026-03-01", want: "invalid --from date"},
		{name: "invalid to", from: "2026-02-01", to: "2026-02-30", want: "invalid --to date"},
		{name: "reversed", from: "2026-03-01", to: "2026-02-28", want: "--from must not be after --to"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"--org-id", "11111111-1111-4111-8111-111111111111", "--from", tc.from, "--to", tc.to}
			err := Command(cfg, requestapi.Organization).Dispatch(args, nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("Dispatch(%v) = %v, want error containing %q", args, err, tc.want)
			}
		})
	}
}

func TestCommandDefaultDatesUseUTC(t *testing.T) {
	original := time.Local
	defer func() { time.Local = original }()
	offset := -12 * 60 * 60
	if time.Now().UTC().Hour() >= 10 {
		offset = 14 * 60 * 60
	}
	time.Local = time.FixedZone("test", offset)
	before := time.Now().UTC()
	cmd := Command(config.Config{}, requestapi.Organization)
	after := time.Now().UTC()
	for _, tc := range []struct {
		flag, before, after string
	}{
		{flag: "to", before: before.Format(time.DateOnly), after: after.Format(time.DateOnly)},
		{flag: "from", before: before.AddDate(0, 0, 1-before.Day()).Format(time.DateOnly), after: after.AddDate(0, 0, 1-after.Day()).Format(time.DateOnly)},
	} {
		got := cmd.Flags.Lookup(tc.flag).DefValue
		if got != tc.before && got != tc.after {
			t.Errorf("default --%s = %q, want %q or %q", tc.flag, got, tc.before, tc.after)
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
	jan := time.Date(2025, time.January, 15, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		period, from, to string
	}{
		{period: "last-month", from: "2024-12-01", to: "2024-12-31"},
		{period: "3-months", from: "2024-11-01", to: "2025-01-15"},
	} {
		t.Run("january/"+tc.period, func(t *testing.T) {
			from, to, err := periodDates(tc.period, jan)
			if err != nil || from != tc.from || to != tc.to {
				t.Errorf("periodDates(%q, %v) = %q, %q, %v; want %q, %q, nil", tc.period, jan, from, to, err, tc.from, tc.to)
			}
		})
	}
	if _, _, err := periodDates("invalid", now); err == nil {
		t.Error("periodDates(invalid) succeeded, want error")
	}
	sydney := time.Date(2024, time.December, 1, 9, 0, 0, 0, time.FixedZone("Sydney", 11*60*60))
	from, to, err := periodDates("this-month", sydney.UTC())
	if err != nil || from != "2024-11-01" || to != "2024-11-30" {
		t.Errorf("periodDates(this-month, %v UTC) = %q, %q, %v; want November 1 to 30", sydney.UTC(), from, to, err)
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

	if err := Command(cfg, requestapi.Organization).Dispatch([]string{"--org-id", "11111111-1111-4111-8111-111111111111", "--period", "last-month", "--json"}, nil); err != nil {
		t.Fatal(err)
	}
	r := <-requests
	from, to, _ := periodDates("last-month", time.Now().UTC())
	if r.URL.Query().Get("from") != from || r.URL.Query().Get("to") != to {
		t.Errorf("request = %s, want from=%s and to=%s", r.URL, from, to)
	}
	for _, flags := range [][]string{{"--period", "invalid"}, {"--period", "this-month", "--from", "2025-01-01"}, {"--period", "this-month", "--to", "2025-01-01"}} {
		if err := Command(cfg, requestapi.Organization).Dispatch(flags, nil); err == nil {
			t.Errorf("Dispatch(%v) succeeded, want error", flags)
		}
	}
}

func TestProjectLeaderboard(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("AppData", dir)
	if err := auth.SaveTokens(auth.TokenCache{AccessToken: "access-token", RefreshToken: "refresh-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/tinacloud.json", []byte(`{"org_id":"22222222-2222-4222-8222-222222222222","project_id":"11111111-1111-4111-8111-111111111111"}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/leaderboard/project/11111111-1111-4111-8111-111111111111" {
			t.Errorf("request path = %q, want configured project leaderboard", r.URL.Path)
		}
		fmt.Fprint(w, `{"entries":[]}`)
	}))
	defer server.Close()
	cfg := config.Config{}
	cfg.TinaCloud.ContentEndpoint = server.URL
	if err := Command(cfg, requestapi.Project).Dispatch([]string{"--tina-dir", ".", "--json"}, nil); err != nil {
		t.Fatal(err)
	}
	if err := Command(cfg, requestapi.Project).Dispatch([]string{"--tina-dir", t.TempDir()}, nil); err == nil || err.Error() != "project ID is required" {
		t.Errorf("Dispatch() = %v, want project ID required", err)
	}
}

func TestProjectLeaderboardLanding(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("AppData", dir)
	if err := auth.SaveTokens(auth.TokenCache{AccessToken: "access-token", RefreshToken: "refresh-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, landing string
		json          bool
	}{
		{name: "text", landing: `{"at":1735689600000,"branch":"feature","contributors":[{"userId":"1","name":"Alice","formerMember":false},{"userId":"2","name":null,"formerMember":true},{"userId":"3","name":null,"formerMember":false}]}`},
		{name: "json", landing: `{"at":1735689600000,"branch":"feature","contributors":[{"userId":"1","name":"Alice","formerMember":false},{"userId":"2","name":null,"formerMember":true},{"userId":"3","name":null,"formerMember":false}]}`, json: true},
		{name: "missing", landing: "null"},
		{name: "missing json", landing: "null", json: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{"entries":[],"landing":%s}`, tc.landing)
			}))
			defer server.Close()
			cfg := config.Config{}
			cfg.TinaCloud.ContentEndpoint = server.URL
			output, err := os.CreateTemp(t.TempDir(), "output")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			stdout := os.Stdout
			os.Stdout = output
			defer func() { os.Stdout = stdout }()
			args := []string{"--project-id", "11111111-1111-4111-8111-111111111111"}
			if tc.json {
				args = append(args, "--json")
			}
			if err := Command(cfg, requestapi.Project).Dispatch(args, nil); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			if tc.json {
				var got map[string]json.RawMessage
				if err := json.Unmarshal(data, &got); err != nil {
					t.Fatal(err)
				}
				var actual, want any
				if tc.landing != "null" {
					if err := json.Unmarshal(got["landing"], &actual); err != nil {
						t.Fatal(err)
					}
					if err := json.Unmarshal([]byte(tc.landing), &want); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(actual, want) {
						t.Errorf("landing = %s, want %s", got["landing"], tc.landing)
					}
				} else if _, ok := got["landing"]; ok {
					t.Errorf("output = %s, want landing omitted", data)
				}
			} else {
				want := "Last landed on default branch at 2025-01-01T00:00:00Z by Alice, Former member, No name set"
				if (tc.landing != "null" && !strings.Contains(string(data), want)) || (tc.landing == "null" && strings.Contains(string(data), "Last landed")) {
					t.Errorf("output = %s, want landing line present = %t", data, tc.landing != "null")
				}
			}
		})
	}
}
