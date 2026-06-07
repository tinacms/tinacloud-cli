package domain

import (
	"fmt"
	"strings"
	"time"
)

type GetV2ProjectResponse struct {
	Tokens    []Token `json:"tokens"`
	OrgID     string  `json:"orgId"`
	ProjectID string  `json:"id"`
	Name      string  `json:"name"`
	RepoURL   string  `json:"repoUrl"`
	// Comma deliminated
	SitURLs           string                     `json:"sitUrl"`
	Metadata          map[Branch]ProjectMetadata `json:"metadata"`
	DefaultBranch     Branch                     `json:"defaultBranch"`
	MediaBranch       Branch                     `json:"mediaBranch"`
	ProtectedBranches *[]Branch                  `json:"protectedBranches"`
	Owner             string                     `json:"owner"`
	Role              string                     `json:"role"`
	// Unix timestamp
	DateCreated int64 `json:"dateCreated"`
	// Unix timestamp
	DateUpdated int64 `json:"dateUpdated"`
	// Unix timestamp
	ContentUpdated int64 `json:"contentUpdated"`
}

func (p GetV2ProjectResponse) String() string {
	var output strings.Builder

	fmt.Fprintf(&output, "%s (%s)\n", p.Name, p.ProjectID)
	fmt.Fprintf(&output, "\t- Organization: %s\n", p.OrgID)
	fmt.Fprintf(&output, "\t- Repo: %s\n", p.RepoURL)
	fmt.Fprintf(&output, "\t- Site URL(s): %s\n", strings.Split(p.SitURLs, ","))
	// Date[Created|Updated] will always contain a valid date as it records the
	// project creation date initially.
	fmt.Fprintf(&output, "\t- Created: %s\n", time.UnixMilli(p.DateCreated).Format(time.RFC822))
	fmt.Fprintf(&output, "\t- Last code change: %s", time.UnixMilli(p.DateUpdated).Format(time.RFC822))

	if len(p.Metadata) > 0 {
		fmt.Fprint(&output, "\n\t- Metadata:")
		for branch, metadata := range p.Metadata {
			fmt.Fprintf(&output, "\n\t\t- %s:", branch)
			fmt.Fprintf(&output, "\n\t\t\t- Tina version: %s", metadata.TinaVersion.String())
		}
	}

	if p.Role != "" {
		fmt.Fprintf(&output, "\n\t- Role: %s", strings.ToTitle(p.Role[0:1])+p.Role[1:])
	}
	if p.DefaultBranch != "" {
		fmt.Fprintf(&output, "\n\t- Default branch: %s", p.DefaultBranch)
	}
	if p.MediaBranch != "" {
		fmt.Fprintf(&output, "\n\t- Media branch: %s", p.MediaBranch)
	}
	if p.ContentUpdated != 0 {
		fmt.Fprintf(&output, "\n\t- Last content change: %s", time.UnixMilli(p.ContentUpdated).Format(time.RFC822))
	}

	return output.String()
}
