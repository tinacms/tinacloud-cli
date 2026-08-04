package status

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/requests"
)

type statusJSONOutput struct {
	LoggedIn bool   `json:"loggedIn"`
	FullName string `json:"fullName,omitempty"`
	Email    string `json:"email,omitempty"`
}

func Status(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("status", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud auth status [flags]")
		fmt.Fprintln(os.Stderr, "\nDisplay authentication status and TinaCloud account information")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	_ = flagSet.Parse(args)

	out, err := runStatus(cfg, auth.TokensFromCache)
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

func runStatus(cfg config.Config, loadTokens auth.TokenLoader) (statusJSONOutput, error) {
	tokens, err := loadTokens(cfg)
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

	if claims.TinaCloudUserID == "" {
		return statusJSONOutput{}, errors.New("failed to retrieve user id from access token")
	}

	user, err := requests.GetTinaCloudUser(cfg.TinaCloud.Endpoint, tokens.AccessToken, claims.TinaCloudUserID)
	if err != nil {
		return statusJSONOutput{}, err
	}

	return statusJSONOutput{LoggedIn: true, FullName: user.FullName, Email: user.Email}, nil
}
