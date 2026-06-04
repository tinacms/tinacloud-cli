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
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
)

type statusJSONOutput struct {
	LoggedIn bool   `json:"loggedIn"`
	FullName string `json:"fullName,omitempty"`
	Email    string `json:"email,omitempty"`
}

func Status(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("status", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli auth status [flags]")
		fmt.Fprintln(os.Stderr, "\nDisplay authentication status and TinaCloud account information")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	_ = flagSet.Parse(args)

	out, err := runStatus(cfg)
	if *jsonFlag {
		return command.WriteJSONResult(os.Stdout, out, err)
	}
	if err != nil {
		return err
	}

	if !out.LoggedIn {
		fmt.Println("You are not currently logged in")
		return nil
	}

	fmt.Printf("Currently signed in as %s (%s)\n", out.FullName, out.Email)
	return nil
}

func runStatus(cfg config.Config) (statusJSONOutput, error) {
	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return statusJSONOutput{LoggedIn: false}, nil
		}
		return statusJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	claims, err := auth.DecodeAccessToken(tokens.AccessToken)
	if err != nil {
		return statusJSONOutput{}, fmt.Errorf("failed to decode access token: %w", err)
	}

	if claims.TinaCloudClientID == "" {
		return statusJSONOutput{}, errors.New("failed to retrieve client id from access token")
	}

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/users/%s", cfg.TinaCloud.Endpoint, claims.TinaCloudClientID), nil)
	if err != nil {
		return statusJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return statusJSONOutput{}, fmt.Errorf("failed to get user information: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return statusJSONOutput{}, fmt.Errorf("failed to read get user information response: %w", err)
	}

	if resp.StatusCode > 299 {
		return statusJSONOutput{}, fmt.Errorf("failed to get user information. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var user domain.TinaCloudUser
	if err := json.Unmarshal(body, &user); err != nil {
		return statusJSONOutput{}, fmt.Errorf("failed to unmarshal tina user from get user request: %w", err)
	}

	return statusJSONOutput{LoggedIn: true, FullName: user.FullName, Email: user.Email}, nil
}
