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
	"tina.io/tinacloud-cli/internal/utils"
)

type statusJSONOutput struct {
	LoggedIn bool   `json:"loggedIn"`
	FullName string `json:"fullName,omitempty"`
	Email    string `json:"email,omitempty"`
}

func Command(cfg config.Config, requestClient *requests.Client) *command.Command {
	flagSet := flag.NewFlagSet("status", flag.ContinueOnError)
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")

	return &command.Command{
		Name:        "status",
		Description: "Display authentication status and TinaCloud account information",
		Flags:       flagSet,
		Run: func(args []string) error {
			out, err := runStatus(cfg, requestClient, auth.TokensFromCache)
			if *jsonFlag {
				return utils.WriteJSONResult(os.Stdout, out, err)
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
		},
	}
}

func runStatus(cfg config.Config, requestClient *requests.Client, loadTokens auth.TokenLoader) (statusJSONOutput, error) {
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

	user, err := requestClient.GetTinaCloudUser(cfg.TinaCloud.Endpoint, tokens.AccessToken, claims.TinaCloudUserID)
	if err != nil {
		return statusJSONOutput{}, err
	}

	return statusJSONOutput{LoggedIn: true, FullName: user.FullName, Email: user.Email}, nil
}
