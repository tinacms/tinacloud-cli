package status

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
)

func Status(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("login", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli auth status")
		fmt.Fprintln(os.Stderr, "\nDisplay authentication status and TinaCloud account information")
	}
	_ = flagSet.Parse(args)

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.UnathenticatedError) {
			fmt.Println("You are not currently logged in")
			return nil
		}
		return fmt.Errorf("failed to retrieve auth tokens: %w", err)
	}

	claims, err := auth.DecodeAccessToken(cfg, tokens.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to decode access token: %w", err)
	}

	if claims.TinaCloudClientID == "" {
		return errors.New("failed to retrieve client id from access token")
	}

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/users/%s", cfg.TinaCloud.Endpoint, claims.TinaCloudClientID), nil)
	if err != nil {
		return fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get user information: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read get user information response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return fmt.Errorf("failed to get user information. receieved response status of '%d': %s", resp.StatusCode, string(body))
	}

	var user domain.TinaCloudUser
	if err := json.Unmarshal(body, &user); err != nil {
		return fmt.Errorf("failed to unmarshal tina user from get user request: %w", err)
	}

	fmt.Printf("Currently signed in as %s (%s)\n", user.FullName, user.Email)

	return nil
}
