package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
)

func GetGitHosts(endpoint, accessToken, orgID string) (domain.GetV2GitHostsResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/organizations/%s/githosts", endpoint, orgID), nil)
	if err != nil {
		return domain.GetV2GitHostsResponse{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := httpclient.Client.Do(req)
	if err != nil {
		return domain.GetV2GitHostsResponse{}, fmt.Errorf("failed to get githosts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.GetV2GitHostsResponse{}, httpclient.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.GetV2GitHostsResponse{}, fmt.Errorf("failed to read get githosts response: %w", err)
	}

	if resp.StatusCode > 299 {
		// They do not have enterprise access
		if resp.StatusCode == http.StatusForbidden {
			return domain.GetV2GitHostsResponse{GitHosts: []domain.GitHost{}}, nil
		}
		return domain.GetV2GitHostsResponse{}, fmt.Errorf("failed to get githosts. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var githosts domain.GetV2GitHostsResponse
	if err := json.Unmarshal(body, &githosts); err != nil {
		return domain.GetV2GitHostsResponse{}, fmt.Errorf("failed to unmarshal githosts from get request: %w", err)
	}

	return githosts, nil
}
