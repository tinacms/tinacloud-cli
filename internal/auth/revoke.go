package auth

import (
	"fmt"
	"io"
	"net/url"

	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/httpclient"
)

// RevokeToken revokes the refresh token upstream, invalidating the session
// and all access tokens issued from it.
func RevokeToken(cfg config.Config, refreshToken string) error {
	params := url.Values{}
	params.Set("token", refreshToken)

	resp, err := httpclient.Client.PostForm(fmt.Sprintf("%s/oauth/revoke", cfg.TinaCloud.IdentityEndpoint), params)
	if err != nil {
		return fmt.Errorf("failed to post token for revocation: %w", httpclient.ClassifyHTTPError(err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token revocation response: %w", httpclient.ClassifyHTTPError(err))
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("failed to revoke token. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	return nil
}
