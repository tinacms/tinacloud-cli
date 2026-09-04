package requests

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/domain"
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
	server, rec := jsonServer(t, http.StatusOK, `[{"organization":{"id":"org-1","name":"Acme"},"role":"admin"}]`)

	got, err := NewClient().GetOrganizations(server.URL, "tok")
	if err != nil {
		t.Fatalf("GetOrganizations() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/organizations" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/organizations")
	}
	if gotAuth := rec.req.Header.Get("Authorization"); gotAuth != "Bearer tok" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer tok")
	}
	if len(got) != 1 || got[0].Organization.ID != "org-1" {
		t.Errorf("GetOrganizations() = %+v, want one org with ID org-1", got)
	}
}

func TestGetOrganizationsErrors(t *testing.T) {
	t.Run("returns error on non-2xx response", func(t *testing.T) {
		server, _ := jsonServer(t, http.StatusInternalServerError, "boom")

		if _, err := NewClient().GetOrganizations(server.URL, "tok"); err == nil {
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
		{"GetOrganizations", func() error { _, err := NewClient().GetOrganizations(server.URL, "tok"); return err }},
		{"GetProjects", func() error { _, err := NewClient().GetProjects(server.URL, "tok", "org-1"); return err }},
		{"GetProject", func() error { _, err := NewClient().GetProject(server.URL, "tok", "project-1"); return err }},
		{"GetTinaCloudUser", func() error { _, err := NewClient().GetTinaCloudUser(server.URL, "tok", "user-1"); return err }},
		{"CreateProject", func() error {
			_, err := NewClient().CreateProject(server.URL, "tok", "org-1", domain.PostV2ProjectsBody{})
			return err
		}},
		{"PutProject", func() error {
			return NewClient().PutProject(server.URL, "tok", "org-1", "project-1", domain.PutV2ProjectBody{})
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, auth.ErrUnauthorized) {
				t.Fatalf("%s = %v, want ErrUnauthorized", tc.name, err)
			}
		})
	}
}

func TestGetProjects(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, `[{"id":"project-1","name":"Site"}]`)

	got, err := NewClient().GetProjects(server.URL, "tok", "org-1")
	if err != nil {
		t.Fatalf("GetProjects() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/organizations/org-1/apps" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/organizations/org-1/apps")
	}
	if len(got) != 1 || got[0].ID != "project-1" {
		t.Errorf("GetProjects() = %+v, want one project with ID project-1", got)
	}
}

func TestGetProject(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, `{"id":"project-1","name":"Site","orgId":"org-1"}`)

	got, err := NewClient().GetProject(server.URL, "tok", "project-1")
	if err != nil {
		t.Fatalf("GetProject() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/apps/project-1" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/apps/project-1")
	}
	if got.ProjectID != "project-1" || got.OrgID != "org-1" {
		t.Errorf("GetProject() = %+v, want project-1 in org-1", got)
	}
}

func TestGetTinaCloudUser(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, `{"email":"a@b.com","fullName":"Ada"}`)

	got, err := NewClient().GetTinaCloudUser(server.URL, "tok", "user-1")
	if err != nil {
		t.Fatalf("GetTinaCloudUser() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/users/user-1" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/users/user-1")
	}
	want := domain.TinaCloudUser{Email: "a@b.com", FullName: "Ada"}
	if got != want {
		t.Errorf("GetTinaCloudUser() = %+v, want %+v", got, want)
	}
}

func TestCreateProject(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, `{"created_app":{"id":"project-1","name":"Site"}}`)

	body := domain.PostV2ProjectsBody{Name: "Site", RepoURL: "https://github.com/a/b", SiteURL: "https://b.com"}
	got, err := NewClient().CreateProject(server.URL, "tok", "org-1", body)
	if err != nil {
		t.Fatalf("CreateProject() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/organizations/org-1/apps" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/organizations/org-1/apps")
	}
	if !strings.Contains(string(rec.body), `"name":"Site"`) {
		t.Errorf("request body = %q, want it to contain the marshalled name", rec.body)
	}
	if got.CreatedProject.ProjectID != "project-1" {
		t.Errorf("CreateProject() = %+v, want created project ID project-1", got)
	}
}

func TestRequestTimeouts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
	}))
	t.Cleanup(server.Close)

	client := &Client{http: &http.Client{Timeout: 10 * time.Millisecond}}
	tests := []struct {
		name      string
		operation string
		call      func() error
	}{
		{name: "GetOrganizations", operation: "get organizations", call: func() error {
			_, err := client.GetOrganizations(server.URL, "tok")
			return err
		}},
		{name: "GetProjects", operation: "get projects", call: func() error {
			_, err := client.GetProjects(server.URL, "tok", "org-1")
			return err
		}},
		{name: "GetProject", operation: "get project", call: func() error {
			_, err := client.GetProject(server.URL, "tok", "project-1")
			return err
		}},
		{name: "GetTinaCloudUser", operation: "get TinaCloud user", call: func() error {
			_, err := client.GetTinaCloudUser(server.URL, "tok", "user-1")
			return err
		}},
		{name: "CreateProject", operation: "create project", call: func() error {
			_, err := client.CreateProject(server.URL, "tok", "org-1", domain.PostV2ProjectsBody{})
			return err
		}},
		{name: "PutProject", operation: "update project", call: func() error {
			return client.PutProject(server.URL, "tok", "org-1", "project-1", domain.PutV2ProjectBody{})
		}},
		{name: "GetGitHosts", operation: "get git hosts", call: func() error {
			_, err := client.GetGitHosts(server.URL, "tok", "org-1")
			return err
		}},
		{name: "GetTinaCloudGitHubApp", operation: "get TinaCloud GitHub app", call: func() error {
			_, err := client.GetTinaCloudGitHubApp(server.URL)
			return err
		}},
		{name: "GetPostHogToken", operation: "get PostHog token", call: func() error {
			_, err := client.GetPostHogToken(server.URL)
			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			t.Logf("timeout error: %v", err)
			if !errors.Is(err, ErrRequestTimeout) {
				t.Fatalf("%s() = %v, want ErrRequestTimeout", tc.name, err)
			}

			want := tc.operation + ": request timed out after 10ms"
			if got := err.Error(); got != want {
				t.Errorf("%s() error = %q, want %q", tc.name, got, want)
			}
		})
	}
}

func TestPutProject(t *testing.T) {
	server, rec := jsonServer(t, http.StatusOK, "")

	body := domain.PutV2ProjectBody{Name: "Renamed"}
	if err := NewClient().PutProject(server.URL, "tok", "org-1", "project-1", body); err != nil {
		t.Fatalf("PutProject() = %v, want nil error", err)
	}
	if rec.req.URL.Path != "/v2/organizations/org-1/apps/project-1" {
		t.Errorf("request path = %q, want %q", rec.req.URL.Path, "/v2/organizations/org-1/apps/project-1")
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

	got, err := NewClient().GetPostHogToken(server.URL)
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
