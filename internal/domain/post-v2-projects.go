package domain

import (
	"fmt"
	"strings"
)

type PostV2ProjectsBody struct {
	Name           string `json:"name"`
	RepoURL        string `json:"repo_url"`
	SiteURL        string `json:"site_url"`
	InstallationID *int64 `json:"installation_id,omitempty"`
}

type PostV2ProjectResponse struct {
	CreatedProject CreatedProject `json:"created_app"`
}

type CreatedProject struct {
	Tokens    []Token `json:"tokens"`
	OrgID     string  `json:"orgId"`
	ProjectID string  `json:"id"`
	Name      string  `json:"name"`
	RepoURL   string  `json:"repoUrl"`
	// Comma deliminated
	SiteURLs string `json:"siteUrl"`
}

func (cp CreatedProject) String() string {
	var output strings.Builder

	fmt.Fprintf(&output, "%s (%s)\n", cp.Name, cp.ProjectID)
	fmt.Fprintf(&output, "\t\t- Organization: %s\n", cp.OrgID)
	fmt.Fprintf(&output, "\t\t- Repo: %s\n", cp.RepoURL)
	fmt.Fprintf(&output, "\t\t- Site URL(s): %s", strings.Split(cp.SiteURLs, ","))

	return output.String()
}
