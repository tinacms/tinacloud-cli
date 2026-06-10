package domain

import (
	"fmt"
	"strings"
	"time"
)

type GetV2ProjectsResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	RepoURL string `json:"repoUrl"`
	// Comma deliminated
	SiteURLs string `json:"siteUrl"`
	// Unix timestamp
	DateUpdated int64 `json:"dateUpdated"`
	// Unix timestamp
	ContentUpdated int64 `json:"contentUpdated"`
}

func (p GetV2ProjectsResponse) String() string {
	var output strings.Builder

	fmt.Fprintf(&output, "%s (%s)\n", p.Name, p.ID)
	fmt.Fprintf(&output, "\t\t- Repo: %s\n", p.RepoURL)
	fmt.Fprintf(&output, "\t\t- URL(s): %s\n", strings.Split(p.SiteURLs, ","))

	// DateUpdated will always contain a valid date as it records the project
	// creation date initially.
	fmt.Fprintf(&output, "\t\t- Last code change: %s", time.UnixMilli(p.DateUpdated).Format(time.RFC822))

	if p.ContentUpdated != 0 {
		fmt.Fprintf(&output, "\n\t\t- Last content change: %s", time.UnixMilli(p.ContentUpdated).Format(time.RFC822))
	}

	return output.String()
}
