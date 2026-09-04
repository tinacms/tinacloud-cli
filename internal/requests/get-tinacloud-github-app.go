package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/domain"
)

func (c *Client) GetTinaCloudGitHubApp(endpoint string) (domain.GetV2TinaCloudGitHubAppResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/github/app", endpoint), nil)
	if err != nil {
		return domain.GetV2TinaCloudGitHubAppResponse{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return domain.GetV2TinaCloudGitHubAppResponse{}, c.requestError("get TinaCloud GitHub app", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.GetV2TinaCloudGitHubAppResponse{}, auth.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.GetV2TinaCloudGitHubAppResponse{}, c.requestError("get TinaCloud GitHub app", err)
	}

	if resp.StatusCode > 299 {
		return domain.GetV2TinaCloudGitHubAppResponse{}, fmt.Errorf("failed to get github app name. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var githubApp domain.GetV2TinaCloudGitHubAppResponse
	if err := json.Unmarshal(body, &githubApp); err != nil {
		return domain.GetV2TinaCloudGitHubAppResponse{}, fmt.Errorf("failed to unmarshal github app name from get request: %w", err)
	}

	return githubApp, nil
}
