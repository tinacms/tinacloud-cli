package domain

import (
	"fmt"
	"strings"
)

type PostV2ProjectsBody struct {
	Name    string `json:"name"`
	RepoURL string `json:"repo_url" `
	SiteURL string `json:"site_url"`
}

type PostV2ProjectResponse struct {
	CreatedProject CreatedProject `json:"created_app"`
}

type Token struct {
	Data     string   `json:"token"`
	Name     string   `json:"name"`
	RepoURL  string   `json:"repoURL"`
	Branches []string `json:"branches"`
	// Unix timestamp
	LastUsed int64    `json:"lastUsed"`
	Access   []string `json:"access"`
}

type CreatedProject struct {
	Tokens    []Token `json:"tokens"`
	OrgID     string  `json:"orgId"`
	ProjectID string  `json:"id"`
	Name      string  `json:"name"`
	RepoURL   string  `json:"repoUrl"`
	// Comma deliminated
	SitURLs string `json:"sitUrl"`
}

func (cp CreatedProject) String() string {
	var output strings.Builder

	fmt.Fprintf(&output, "%s (%s)\n", cp.Name, cp.ProjectID)
	fmt.Fprintf(&output, "\t\t- Organization: %s\n", cp.OrgID)
	fmt.Fprintf(&output, "\t\t- Repo: %s\n", cp.RepoURL)
	fmt.Fprintf(&output, "\t\t- Site URL(s): %s", strings.Split(cp.SitURLs, ","))

	return output.String()
}
