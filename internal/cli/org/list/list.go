package list

import (
	"errors"
	"flag"
	"fmt"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
	"tina.io/tinacloud-cli/internal/requests"
	"tina.io/tinacloud-cli/internal/utils"
)

type listJSONOutput struct {
	LoggedIn      bool                          `json:"loggedIn"`
	Organizations []domain.OrganizationWithRole `json:"organizations"`
}

func Command(cfg config.Config) *command.Command {
	flagSet := flag.NewFlagSet("list", flag.ContinueOnError)
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")

	return &command.Command{
		Name:        "list",
		Description: "List organizations",
		Flags:       flagSet,
		Run: func(args []string) error {
			out, err := runList(cfg, auth.TokensFromCache)
			if *jsonFlag {
				return utils.WriteJSONResult(out, err)
			}
			if err != nil {
				return err
			}

			fmt.Println("Organizations:")
			for _, org := range out.Organizations {
				fmt.Printf("\t- %s\n", org.String())
			}

			return nil
		},
	}
}

func runList(cfg config.Config, loadTokens auth.TokenLoader) (listJSONOutput, error) {
	tokens, err := loadTokens(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, httpclient.ErrUnauthorized) {
			return listJSONOutput{}, err
		}
		return listJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	organizations, err := requests.GetOrganizations(cfg.TinaCloud.IdentityEndpoint, tokens.AccessToken)
	if err != nil {
		return listJSONOutput{}, err
	}

	return listJSONOutput{
		LoggedIn:      true,
		Organizations: organizations,
	}, nil
}
