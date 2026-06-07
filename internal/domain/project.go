package domain

import (
	"fmt"
	"strings"
	"time"
)

type Project struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	RepoURL        string    `json:"repo_url"`
	SiteURLs       []string  `json:"site_urls"`
	DateUpdated    time.Time `json:"date_updated"`
	ContentUpdated time.Time `json:"content_updated"`
}

func (p Project) String() string {
	var output strings.Builder

	fmt.Fprintf(&output, "%s (%s)\n", p.Name, p.ID)
	fmt.Fprintf(&output, "\t\t- Repo: %s\n", p.RepoURL)
	fmt.Fprintf(&output, "\t\t- URLs: %s\n", p.SiteURLs)
	fmt.Fprintf(&output, "\t\t- Last code change: %s\n", p.DateUpdated.Format(time.RFC822))

	// DateUpdated will always contain a valid date as it records the project
	// creation date initially. ContentUpdated may be 0 if no content has been
	// pushed yet.
	if p.ContentUpdated.UnixMilli() != 0 {
		fmt.Fprintf(&output, "\t- Last content change: %s\n", p.ContentUpdated.Format(time.RFC822))
	}

	return output.String()
}
