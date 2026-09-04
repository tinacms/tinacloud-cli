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
	"tina.io/tinacloud-cli/internal/utils"
)

type statusJSONOutput struct {
	LoggedIn bool                        `json:"loggedIn"`
	Project  domain.GetV2ProjectResponse `json:"project"`
}

func Command(cfg config.Config, requestClient *requests.Client) *command.Command {
	flagSet := flag.NewFlagSet("list", flag.ContinueOnError)
	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")

	return &command.Command{
		Name:        "status",
		Description: "List a project's status",
		Flags:       flagSet,
		Run: func(args []string) error {
			out, err := runStatus(cfg, requestClient, *projectIDFlag, *tinaDirFlag)
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

			fmt.Printf("%s\n", out.Project.String())

			return nil
		},
	}
}

func runStatus(cfg config.Config, requestClient *requests.Client, projectID, tinaDir string) (statusJSONOutput, error) {
	if projectID == "" {
		if cwd, err := os.Getwd(); err == nil {
			if tinacloudConfig, err := domain.TinaCloudConfigFromDisk(cwd, tinaDir); err == nil {
				projectID = tinacloudConfig.ProjectID
			}
		}

		if projectID == "" {
			return statusJSONOutput{}, errors.New("project ID is required")
		}
	}

	tokens, err := auth.TokensFromCache(cfg, requestClient)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return statusJSONOutput{LoggedIn: false}, nil
		}
		return statusJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	project, err := requestClient.GetProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, projectID)
	if err != nil {
		return statusJSONOutput{}, err
	}

	return statusJSONOutput{
		LoggedIn: true,
		Project:  project,
	}, nil
}
