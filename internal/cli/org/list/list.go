package list

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/requests"
)

type listJSONOutput struct {
	LoggedIn      bool                          `json:"loggedIn"`
	Organizations []domain.OrganizationWithRole `json:"organizations"`
}

func List(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("orglist", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud org list [flags]")
		fmt.Fprintln(os.Stderr, "\nList organizations")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	_ = flagSet.Parse(args)

	out, err := runList(cfg, auth.TokensFromCache)
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

	fmt.Println("Organizations:")
	for _, org := range out.Organizations {
		fmt.Printf("\t- %s\n", org.String())
	}

	return nil
}

func runList(cfg config.Config, loadTokens auth.TokenLoader) (listJSONOutput, error) {
	tokens, err := loadTokens(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return listJSONOutput{LoggedIn: false}, nil
		}
		return listJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	organizations, err := requests.GetOrganizations(cfg.TinaCloud.Endpoint, tokens.AccessToken)
	if err != nil {
		return listJSONOutput{}, err
	}

	return listJSONOutput{
		LoggedIn:      true,
		Organizations: organizations,
	}, nil
}
