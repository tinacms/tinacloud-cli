package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/domain"
)

// RefreshToken exchanges a refresh token through the TinaCloud OAuth endpoint.
func (c *Client) RefreshToken(endpoint, refreshToken string) (domain.AuthResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	req, err := http.NewRequest(http.MethodPost, endpoint+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("create token refresh request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return domain.AuthResponse{}, c.requestError("refresh auth token", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.AuthResponse{}, auth.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.AuthResponse{}, c.requestError("refresh auth token", err)
	}

	if resp.StatusCode > 299 {
		return domain.AuthResponse{}, fmt.Errorf("refresh auth token: response status %d: %s", resp.StatusCode, string(body))
	}

	var tokens domain.AuthResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return domain.AuthResponse{}, fmt.Errorf("decode refreshed tokens: %w", err)
	}

	return tokens, nil
}
