package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
)

func GetTinaCloudUser(endpoint, accessToken, userID string) (domain.TinaCloudUser, error) {
	if err := validateIDs(userID); err != nil {
		return domain.TinaCloudUser{}, err
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/users/%s", endpoint, userID), nil)
	if err != nil {
		return domain.TinaCloudUser{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := httpclient.Client.Do(req)
	if err != nil {
		return domain.TinaCloudUser{}, fmt.Errorf("failed to get user information: %w", httpclient.ClassifyHTTPError(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.TinaCloudUser{}, httpclient.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.TinaCloudUser{}, fmt.Errorf("failed to read get user information response: %w", httpclient.ClassifyHTTPError(err))
	}

	if resp.StatusCode > 299 {
		return domain.TinaCloudUser{}, fmt.Errorf("failed to get user information. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var user domain.TinaCloudUser
	if err := json.Unmarshal(body, &user); err != nil {
		return domain.TinaCloudUser{}, fmt.Errorf("failed to unmarshal tina user from get user request: %w", err)
	}

	return user, nil
}
