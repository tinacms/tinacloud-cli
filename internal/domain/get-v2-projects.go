package domain

type GetV2Projects struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	RepoURL string `json:"repoUrl"`
	// Comma deliminated
	SiteURLs string `json:"siteUrl"`
	// Unix time-stamp
	DateUpdated int64 `json:"dateUpdated"`
	// Unix time-stamp
	ContentUpdated int64 `json:"contentUpdated"`
}
