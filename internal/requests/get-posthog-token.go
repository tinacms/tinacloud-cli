package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/domain"
)

func GetPostHogToken(endpoint string) (domain.GetV2PostHogTokenResponse, error) {
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/posthog-token", endpoint), nil)
	if err != nil {
		return domain.GetV2PostHogTokenResponse{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return domain.GetV2PostHogTokenResponse{}, fmt.Errorf("failed to get posthog token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.GetV2PostHogTokenResponse{}, auth.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.GetV2PostHogTokenResponse{}, fmt.Errorf("failed to read get posthog token response: %w", err)
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
