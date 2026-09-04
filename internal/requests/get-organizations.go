package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/domain"
)

func (c *Client) GetOrganizations(endpoint, accessToken string) ([]domain.OrganizationWithRole, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/organizations", endpoint), nil)
	if err != nil {
		return []domain.OrganizationWithRole{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return []domain.OrganizationWithRole{}, c.requestError("get organizations", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return []domain.OrganizationWithRole{}, auth.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []domain.OrganizationWithRole{}, c.requestError("get organizations", err)
	}

	if resp.StatusCode > 299 {
		return []domain.OrganizationWithRole{}, fmt.Errorf("failed to get organizations. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var organizations []domain.OrganizationWithRole
	if err := json.Unmarshal(body, &organizations); err != nil {
		return []domain.OrganizationWithRole{}, fmt.Errorf("failed to unmarshal organizations from get request: %w", err)
	}

	return organizations, nil
}
