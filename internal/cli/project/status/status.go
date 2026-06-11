package status

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

type statusJSONOutput struct {
	LoggedIn bool                        `json:"loggedIn"`
	Project  domain.GetV2ProjectResponse `json:"project"`
}

func Status(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("projectstatus", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli project status [flags]")
		fmt.Fprintln(os.Stderr, "\nList a project's status")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	_ = flagSet.Parse(args)

	out, err := runStatus(cfg, *projectIDFlag)
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

	fmt.Printf("%s\n", out.Project.String())

	return nil
}

func runStatus(cfg config.Config, projectID string) (statusJSONOutput, error) {
	if projectID == "" {
		return statusJSONOutput{}, fmt.Errorf("project ID is required")
	}

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return statusJSONOutput{LoggedIn: false}, nil
		}
		return statusJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	project, err := requests.GetProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, projectID)
	if err != nil {
		return statusJSONOutput{}, err
	}

	return statusJSONOutput{
		LoggedIn: true,
		Project:  project,
	}, nil
}
