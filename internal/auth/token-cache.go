package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
)

var ErrUnauthenticated = errors.New("not logged in")
var ErrUnauthorized = errors.New("no access")

// TokenRefresher exchanges a refresh token for a new set of authentication tokens.
type TokenRefresher interface {
	RefreshToken(endpoint, refreshToken string) (domain.AuthResponse, error)
}

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

// NewTokenLoader returns a token loader that refreshes expired tokens with refresher.
func NewTokenLoader(refresher TokenRefresher) TokenLoader {
	return func(cfg config.Config) (TokenCache, error) {
		return TokensFromCache(cfg, refresher)
	}
}

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

func TokensFromCache(cfg config.Config, refresher TokenRefresher) (TokenCache, error) {
	tokens, err := PeekTokens()
	if err != nil {
		return TokenCache{}, err
	}

	if tokens.ExpiresAt.Before(time.Now()) {
		fmt.Fprintln(os.Stderr, "auth token expired, refreshing...")
		authTokens, err := refresher.RefreshToken(cfg.TinaCloud.Endpoint, tokens.RefreshToken)
		if err != nil {
			return TokenCache{}, err
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
		fmt.Fprintln(os.Stderr, "successfully refreshed auth token")
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
