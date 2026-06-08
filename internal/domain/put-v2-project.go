package domain

type PutV2ProjectBody struct {
	Name              string    `json:"name"`
	RepoURL           string    `json:"repo_url"`
	InstallationID    int       `json:"installation_id"`
	PathToTina        string    `json:"path_to_tina"`
	ProtectedBranches *[]Branch `json:"protected_branches"`
	Features          *[]string `json:"features"`
	// Comma deliminated
	SitURL string `json:"site_url"`
}
