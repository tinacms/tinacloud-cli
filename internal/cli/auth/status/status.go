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

type statusOutput struct {
	LoggedIn bool   `json:"loggedIn"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
}

func Status(args []string, cfg config.Config) (err error) {
	flagSet := flag.NewFlagSet("status", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli auth status [flags]")
		fmt.Fprintln(os.Stderr, "\nDisplay authentication status and TinaCloud account information")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var JSON = flagSet.Bool("json", false, "Print the result as JSON")
	_ = flagSet.Parse(args)

	// Under --json, surface errors as JSON on stdout. The error is still
	// returned so the process exits non-zero.
	defer func() {
		if err == nil || !*JSON {
			return
		}
		if data, mErr := json.Marshal(struct {
			Error string `json:"error"`
		}{err.Error()}); mErr == nil {
			fmt.Println(string(data))
		}
	}()

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) {
			if *JSON {
				data, err := json.Marshal(statusOutput{LoggedIn: false})
				if err != nil {
					return fmt.Errorf("failed to marshal status output: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}
			fmt.Println("You are not currently logged in")
			return nil
		}
		return fmt.Errorf("failed to retrieve auth tokens: %w", err)
	}

	claims, err := auth.DecodeAccessToken(tokens.AccessToken)
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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read get user information response: %w", err)
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("failed to get user information. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var user domain.TinaCloudUser
	if err := json.Unmarshal(body, &user); err != nil {
		return fmt.Errorf("failed to unmarshal tina user from get user request: %w", err)
	}

	if *JSON {
		data, err := json.Marshal(statusOutput{LoggedIn: true, Name: user.FullName, Email: user.Email})
		if err != nil {
			return fmt.Errorf("failed to marshal status output: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Currently signed in as %s (%s)\n", user.FullName, user.Email)
	return nil
}
