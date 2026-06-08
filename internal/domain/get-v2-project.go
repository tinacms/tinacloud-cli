package domain

import (
	"fmt"
	"sort"
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
	SiteURLs          string                     `json:"siteUrl"`
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
	if p.Owner != "" {
		fmt.Fprintf(&output, "\t- Owner: %s\n", p.Owner)
	}
	fmt.Fprintf(&output, "\t- Repo: %s\n", p.RepoURL)
	fmt.Fprintf(&output, "\t- Site URL(s): %s\n", strings.Split(p.SiteURLs, ","))
	// Date[Created|Updated] will always contain a valid date as it records the
	// project creation date initially.
	fmt.Fprintf(&output, "\t- Created: %s\n", time.UnixMilli(p.DateCreated).Format(time.RFC822))
	fmt.Fprintf(&output, "\t- Last code change: %s", time.UnixMilli(p.DateUpdated).Format(time.RFC822))

	if len(p.Metadata) > 0 {
		fmt.Fprint(&output, "\n\t- Metadata:")
		branches := make([]Branch, 0, len(p.Metadata))
		for branch := range p.Metadata {
			branches = append(branches, branch)
		}
		sort.Slice(branches, func(i, j int) bool { return branches[i] < branches[j] })
		for _, branch := range branches {
			fmt.Fprintf(&output, "\n\t\t- %s:", branch)
			fmt.Fprintf(&output, "\n\t\t\t- Tina version: %s", p.Metadata[branch].TinaVersion.String())
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
	if p.ProtectedBranches != nil && len(*p.ProtectedBranches) > 0 {
		fmt.Fprintf(&output, "\n\t- Protected branches: %s", *p.ProtectedBranches)
	}
	if p.ContentUpdated != 0 {
		fmt.Fprintf(&output, "\n\t- Last content change: %s", time.UnixMilli(p.ContentUpdated).Format(time.RFC822))
	}

	if len(p.Tokens) > 0 {
		fmt.Fprint(&output, "\n\t- Tokens:")
		for _, token := range p.Tokens {
			fmt.Fprintf(&output, "\n\t\t- %s: %s (%s)", token.Name, token.Data, strings.Join(token.Access, ", "))
		}
	}

	return output.String()
}
