package requests

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
)

type recorder struct {
	req  *http.Request
	body []byte
}

// jsonServer returns a test server that records the last request (and its body)
// and replies with the given status and body.
func jsonServer(t *testing.T, status int, body string) (*httptest.Server, *recorder) {
	t.Helper()
	rec := &recorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.req = r
		rec.body, _ = io.ReadAll(r.Body)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(server.Close)
	return server, rec
}

func TestGetOrganizations(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, `[{"organization":{"id":"11111111-1111-4111-8111-111111111111","name":"Acme"},"role":"admin"}]`)

	got, err := GetOrganizations(server.URL, "tok")
	if err != nil {
		t.Fatalf("GetOrganizations() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/organizations" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/organizations")
	}
	if gotAuth := rec.req.Header.Get("Authorization"); gotAuth != "Bearer tok" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer tok")
	}
	if len(got) != 1 || got[0].Organization.ID != "11111111-1111-4111-8111-111111111111" {
		t.Errorf("GetOrganizations() = %+v, want one org with ID 11111111-1111-4111-8111-111111111111", got)
	}
}

func TestGetOrganizationsErrors(t *testing.T) {
	t.Run("returns error on non-2xx response", func(t *testing.T) {
		server, _ := jsonServer(t, http.StatusInternalServerError, "boom")

		if _, err := GetOrganizations(server.URL, "tok"); err == nil {
			t.Fatal("GetOrganizations() = nil error, want error")
		}
	})
}

// Each request function maps a 401 response to ErrUnauthorized independently,
// so each needs its own coverage.
func TestRequestsMap401ToErrUnauthorized(t *testing.T) {
	server, _ := jsonServer(t, http.StatusUnauthorized, "")

	tests := []struct {
		name string
		call func() error
	}{
		{"GetOrganizations", func() error { _, err := GetOrganizations(server.URL, "tok"); return err }},
		{"GetProjects", func() error {
			_, err := GetProjects(server.URL, "tok", "11111111-1111-4111-8111-111111111111")
			return err
		}},
		{"GetProject", func() error {
			_, err := GetProject(server.URL, "tok", "22222222-2222-4222-9222-222222222222")
			return err
		}},
		{"GetTinaCloudUser", func() error {
			_, err := GetTinaCloudUser(server.URL, "tok", "33333333-3333-4333-a333-333333333333")
			return err
		}},
		{"CreateProject", func() error {
			_, err := CreateProject(server.URL, "tok", "11111111-1111-4111-8111-111111111111", domain.PostV2ProjectsBody{})
			return err
		}},
		{"PutProject", func() error {
			return PutProject(server.URL, "tok", "11111111-1111-4111-8111-111111111111", "22222222-2222-4222-9222-222222222222", domain.PutV2ProjectBody{})
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, httpclient.ErrUnauthorized) {
				t.Fatalf("%s = %v, want ErrUnauthorized", tc.name, err)
			}
		})
	}
}

func TestGetProjects(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, `[{"id":"22222222-2222-4222-9222-222222222222","name":"Site"}]`)

	got, err := GetProjects(server.URL, "tok", "11111111-1111-4111-8111-111111111111")
	if err != nil {
		t.Fatalf("GetProjects() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/organizations/11111111-1111-4111-8111-111111111111/apps" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/organizations/11111111-1111-4111-8111-111111111111/apps")
	}
	if len(got) != 1 || got[0].ID != "22222222-2222-4222-9222-222222222222" {
		t.Errorf("GetProjects() = %+v, want one project with ID 22222222-2222-4222-9222-222222222222", got)
	}
}

func TestGetProject(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, `{"id":"22222222-2222-4222-9222-222222222222","name":"Site","orgId":"11111111-1111-4111-8111-111111111111"}`)

	got, err := GetProject(server.URL, "tok", "22222222-2222-4222-9222-222222222222")
	if err != nil {
		t.Fatalf("GetProject() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/apps/22222222-2222-4222-9222-222222222222" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/apps/22222222-2222-4222-9222-222222222222")
	}
	if got.ProjectID != "22222222-2222-4222-9222-222222222222" || got.OrgID != "11111111-1111-4111-8111-111111111111" {
		t.Errorf("GetProject() = %+v, want 22222222-2222-4222-9222-222222222222 in 11111111-1111-4111-8111-111111111111", got)
	}
}

func TestGetTinaCloudUser(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, `{"email":"a@b.com","fullName":"Ada"}`)

	got, err := GetTinaCloudUser(server.URL, "tok", "33333333-3333-4333-a333-333333333333")
	if err != nil {
		t.Fatalf("GetTinaCloudUser() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/users/33333333-3333-4333-a333-333333333333" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/users/33333333-3333-4333-a333-333333333333")
	}
	want := domain.TinaCloudUser{Email: "a@b.com", FullName: "Ada"}
	if got != want {
		t.Errorf("GetTinaCloudUser() = %+v, want %+v", got, want)
	}
}

func TestCreateProject(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, `{"created_app":{"id":"22222222-2222-4222-9222-222222222222","name":"Site"}}`)

	body := domain.PostV2ProjectsBody{Name: "Site", RepoURL: "https://github.com/a/b", SiteURL: "https://b.com"}
	got, err := CreateProject(server.URL, "tok", "11111111-1111-4111-8111-111111111111", body)
	if err != nil {
		t.Fatalf("CreateProject() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/organizations/11111111-1111-4111-8111-111111111111/apps" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/organizations/11111111-1111-4111-8111-111111111111/apps")
	}
	if !strings.Contains(string(rec.body), `"name":"Site"`) {
		t.Errorf("request body = %q, want it to contain the marshalled name", rec.body)
	}
	if got.CreatedProject.ProjectID != "22222222-2222-4222-9222-222222222222" {
		t.Errorf("CreateProject() = %+v, want created project ID 22222222-2222-4222-9222-222222222222", got)
	}
}

func TestPutProject(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, "")

	body := domain.PutV2ProjectBody{Name: "Renamed"}
	if err := PutProject(server.URL, "tok", "11111111-1111-4111-8111-111111111111", "22222222-2222-4222-9222-222222222222", body); err != nil {
		t.Fatalf("PutProject() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/organizations/11111111-1111-4111-8111-111111111111/apps/22222222-2222-4222-9222-222222222222" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/organizations/11111111-1111-4111-8111-111111111111/apps/22222222-2222-4222-9222-222222222222")
	}
	if rec.req.Method != http.MethodPut {
		t.Errorf("request method = %q, want %q", rec.req.Method, http.MethodPut)
	}
	if !strings.Contains(string(rec.body), `"name":"Renamed"`) {
		t.Errorf("request body = %q, want it to contain the marshalled name", rec.body)
	}
}

func TestGetPostHogToken(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, `{"api_key":"1234","endpoint":"https://ph.example"}`)

	got, err := GetPostHogToken(server.URL)
	if err != nil {
		t.Fatalf("GetPostHogToken() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/posthog-token" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/posthog-token")
	}
	if got.APIKey != "1234" || got.Endpoint != "https://ph.example" {
		t.Errorf("GetPostHogToken() = %+v, want api-key: 1234; endpoint: https://ph.example", got)
	}
}

func TestRequestsRejectInvalidIDs(t *testing.T) {
	const validID = "11111111-1111-4111-8111-111111111111"
	calls := map[string]func(string) error{
		"projects":       func(id string) error { _, err := GetProjects("", "", id); return err },
		"project":        func(id string) error { _, err := GetProject("", "", id); return err },
		"user":           func(id string) error { _, err := GetTinaCloudUser("", "", id); return err },
		"hosts":          func(id string) error { _, err := GetGitHosts("", "", id); return err },
		"create":         func(id string) error { _, err := CreateProject("", "", id, domain.PostV2ProjectsBody{}); return err },
		"update org":     func(id string) error { return PutProject("", "", id, validID, domain.PutV2ProjectBody{}) },
		"update project": func(id string) error { return PutProject("", "", validID, id, domain.PutV2ProjectBody{}) },
		"leaderboard":    func(id string) error { _, err := GetLeaderboard("", "", id, "", ""); return err },
	}
	for name, call := range calls {
		for _, id := range []string{"", "..", "../other", validID + "?admin=true", validID + "%2fother", "11111111-1111-1111-1111-11111111111g"} {
			t.Run(name+"/"+id, func(t *testing.T) {
				if err := call(id); err == nil || !strings.Contains(err.Error(), "invalid UUID") {
					t.Errorf("call(%q) = %v, want invalid UUID error", id, err)
				}
			})
		}
	}
}

func TestValidateIDsUUIDv4(t *testing.T) {
	for _, id := range []string{
		"abcdefab-1234-4234-8234-abcdefabcdef",
		"ABCDEFAB-1234-4234-9234-ABCDEFABCDEF",
		"abcdefab-1234-4234-a234-abcdefabcdef",
		"ABCDEFAB-1234-4234-B234-ABCDEFABCDEF",
	} {
		if err := validateIDs(id); err != nil {
			t.Errorf("validateIDs(%q) = %v, want nil", id, err)
		}
	}
	for _, id := range []string{
		"abcdefab-1234-1234-8234-abcdefabcdef",
		"abcdefab-1234-4234-c234-abcdefabcdef",
		"{abcdefab-1234-4234-8234-abcdefabcdef}",
		"urn:uuid:abcdefab-1234-4234-8234-abcdefabcdef",
		"abcdefab123442348234abcdefabcdef",
	} {
		if err := validateIDs(id); err == nil {
			t.Errorf("validateIDs(%q) = nil, want error", id)
		}
	}
}

func TestGetLeaderboardEncodesQuery(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, `{"entries":[]}`)
	_, err := GetLeaderboard(server.URL, "token", "ABCDEFAB-1234-4234-B234-ABCDEFABCDEF", "today&extra=1", "start#fragment")
	if err != nil {
		t.Fatal(err)
	}
	query := rec.req.URL.Query()
	if query.Get("to") != "today&extra=1" || query.Get("from") != "start#fragment" || query.Has("extra") {
		t.Errorf("query = %v, want literal to and from values", query)
	}
}
