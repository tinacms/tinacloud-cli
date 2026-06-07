package domain

type Token struct {
	Data     string   `json:"token"`
	Name     string   `json:"name"`
	RepoURL  string   `json:"repoURL"`
	Branches []string `json:"branches"`
	// Unix timestamp
	LastUsed int64    `json:"lastUsed"`
	Access   []string `json:"access"`
}
