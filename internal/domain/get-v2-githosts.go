package domain

type GitHost struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	HostType          string `json:"host_type"`
	Server            string `json:"server"`
	GitHubAppName     string `json:"github_app_name"`
	GitHubAppClientID string `json:"github_app_client_id"`
}

type GetV2GitHostsResponse struct {
	GitHosts []GitHost `json:"git_hosts"`
}
