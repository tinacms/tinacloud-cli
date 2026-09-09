package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
)

func GetTinaCloudGitHubApp(endpoint string) (domain.GetV2TinaCloudGitHubAppResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/github/app", endpoint), nil)
	if err != nil {
		return domain.GetV2TinaCloudGitHubAppResponse{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpclient.Client.Do(req)
	if err != nil {
		return domain.GetV2TinaCloudGitHubAppResponse{}, fmt.Errorf("failed to get github app name: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.GetV2TinaCloudGitHubAppResponse{}, httpclient.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.GetV2TinaCloudGitHubAppResponse{}, fmt.Errorf("failed to read get github app name response: %w", err)
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
