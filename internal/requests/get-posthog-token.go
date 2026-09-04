package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tina.io/tinacloud-cli/internal/domain"
)

func (c *Client) GetPostHogToken(endpoint string) (domain.GetV2PostHogTokenResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/posthog-token", endpoint), nil)
	if err != nil {
		return domain.GetV2PostHogTokenResponse{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return domain.GetV2PostHogTokenResponse{}, c.requestError("get PostHog token", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.GetV2PostHogTokenResponse{}, c.requestError("get PostHog token", err)
	}

	if resp.StatusCode > 299 {
		return domain.GetV2PostHogTokenResponse{}, fmt.Errorf("failed to get posthog token. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var token domain.GetV2PostHogTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return domain.GetV2PostHogTokenResponse{}, fmt.Errorf("failed to unmarshal posthog token from get request: %w", err)
	}

	return token, nil
}
