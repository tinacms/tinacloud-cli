package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
)

var ErrUnauthenticated = errors.New("not logged in")

type TokenCache struct {
	IDToken      string    `json:"id_token"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func cachePath() (string, error) {
	configPath, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to read config directory as $HOME and/or $XDG_CONFIG_HOME not set: %w", err)
	}

	return path.Join(configPath, "tinacloud", "cache", "tokens.json"), nil
}

func SaveTokens(tokens TokenCache) error {
	configPath, err := cachePath()
	if err != nil {
		return fmt.Errorf("failed to retrieve config path: %w", err)
	}

	err = os.MkdirAll(path.Dir(configPath), 0700)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.Marshal(tokens) // #nosec G117
	if err != nil {
		return fmt.Errorf("failed to marshal auth tokens: %w", err)
	}

	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write tokens to config cache: %w", err)
	}

	return nil
}

// Load authentication tokens.
type TokenLoader func(config.Config) (TokenCache, error)

// Read cached tokens without refreshing expired access tokens.
func PeekTokens() (TokenCache, error) {
	configPath, err := cachePath()
	if err != nil {
		return TokenCache{}, fmt.Errorf("failed to retrieve config path: %w", err)
	}

	file, err := os.OpenInRoot(path.Dir(configPath), path.Base(configPath))
	if err != nil {
		if os.IsNotExist(err) {
			return TokenCache{}, ErrUnauthenticated
		}
		return TokenCache{}, fmt.Errorf("failed to read token cache as root: %w", err)
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		return TokenCache{}, fmt.Errorf("failed to read token cache: %w", err)
	}

	var tokens TokenCache
	if err := json.Unmarshal(bytes, &tokens); err != nil {
		return TokenCache{}, fmt.Errorf("failed to parse token cache: %w", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		return TokenCache{}, ErrUnauthenticated
	}

	return tokens, nil
}

func TokensFromCache(cfg config.Config) (TokenCache, error) {
	tokens, err := PeekTokens()
	if err != nil {
		return TokenCache{}, err
	}

	if tokens.ExpiresAt.Before(time.Now()) {
		params := url.Values{}
		params.Set("grant_type", "refresh_token")
		params.Set("refresh_token", tokens.RefreshToken)

		resp, err := httpclient.Client.PostForm(fmt.Sprintf("%s/oauth/token", cfg.TinaCloud.IdentityEndpoint), params)
		if err != nil {
			return TokenCache{}, fmt.Errorf("failed to post code to cognito for refresh: %w", httpclient.ClassifyHTTPError(err))
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return TokenCache{}, httpclient.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return TokenCache{}, fmt.Errorf("failed to read auth refresh response: %w", httpclient.ClassifyHTTPError(err))
		}

		if resp.StatusCode > 299 {
			return TokenCache{}, fmt.Errorf("failed to refresh auth token. received response status of '%d': %s", resp.StatusCode, string(body))
		}

		var authTokens domain.AuthResponse
		if err := json.Unmarshal(body, &authTokens); err != nil {
			return TokenCache{}, fmt.Errorf("could not parse tokens: %w", err)
		}

		tokens = TokenCache{
			IDToken:      authTokens.IDToken,
			AccessToken:  authTokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			TokenType:    authTokens.TokenType,
			ExpiresAt:    time.Now().Add(time.Duration(authTokens.ExpiresIn) * time.Second),
		}

		if err := SaveTokens(tokens); err != nil {
			return TokenCache{}, fmt.Errorf("failed to save tokens to file: %w", err)
		}
	}

	return tokens, nil
}

func ClearTokens() error {
	configPath, err := cachePath()
	if err != nil {
		return fmt.Errorf("failed to retrieve config path: %w", err)
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete tokens file: %w", err)
	}

	return nil
}
