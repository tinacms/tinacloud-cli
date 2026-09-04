package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/domain"
)

func (c *Client) PutProject(endpoint, accessToken, orgID, projectID string, body domain.PutV2ProjectBody) error {
	putBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v2/organizations/%s/apps/%s", endpoint, orgID, projectID), bytes.NewBuffer(putBody))
	if err != nil {
		return fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return c.requestError("update project", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return auth.ErrUnauthorized
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.requestError("update project", err)
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("failed to update project. received response status of '%d': %s", resp.StatusCode, string(respBody))
	}

	return nil
}
