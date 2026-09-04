package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/domain"
)

func (c *Client) GetProjects(endpoint, accessToken, orgID string) ([]domain.GetV2ProjectsResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/organizations/%s/apps", endpoint, orgID), nil)
	if err != nil {
		return []domain.GetV2ProjectsResponse{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return []domain.GetV2ProjectsResponse{}, c.requestError("get projects", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return []domain.GetV2ProjectsResponse{}, auth.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []domain.GetV2ProjectsResponse{}, c.requestError("get projects", err)
	}

	if resp.StatusCode > 299 {
		return []domain.GetV2ProjectsResponse{}, fmt.Errorf("failed to get projects. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var projects []domain.GetV2ProjectsResponse
	if err := json.Unmarshal(body, &projects); err != nil {
		return []domain.GetV2ProjectsResponse{}, fmt.Errorf("failed to unmarshal projects from get request: %w", err)
	}

	return projects, nil
}
