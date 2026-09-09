package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
)

func GetProject(endpoint, accessToken, projectID string) (domain.GetV2ProjectResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/apps/%s", endpoint, projectID), nil)
	if err != nil {
		return domain.GetV2ProjectResponse{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := httpclient.Client.Do(req)
	if err != nil {
		return domain.GetV2ProjectResponse{}, fmt.Errorf("failed to get project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.GetV2ProjectResponse{}, httpclient.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.GetV2ProjectResponse{}, fmt.Errorf("failed to read get project response: %w", err)
	}

	if resp.StatusCode > 299 {
		return domain.GetV2ProjectResponse{}, fmt.Errorf("failed to get project. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var project domain.GetV2ProjectResponse
	if err := json.Unmarshal(body, &project); err != nil {
		return domain.GetV2ProjectResponse{}, fmt.Errorf("failed to unmarshal project from get request: %w", err)
	}

	return project, nil
}
