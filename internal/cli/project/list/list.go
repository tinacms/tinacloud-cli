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
	"tina.io/tinacloud-cli/internal/httpclient"
	"tina.io/tinacloud-cli/internal/requests"
	"tina.io/tinacloud-cli/internal/utils"
)

type listJSONOutput struct {
	LoggedIn bool                           `json:"loggedIn"`
	Projects []domain.GetV2ProjectsResponse `json:"projects"`
}

func Command(cfg config.Config) *command.Command {
	flagSet := flag.NewFlagSet("list", flag.ContinueOnError)
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")

	return &command.Command{
		Name:        "list",
		Description: "List projects",
		Example:     "tinacloud project list --org-id 1e1d3398-8255-4812-9f67-388993734a01",
		Flags:       flagSet,
		Run: func(args []string) error {
			out, err := runList(cfg, *orgIDFlag, *tinaDirFlag)
			if *jsonFlag {
				return utils.WriteJSONResult(out, err)
			}
			if err != nil {
				return err
			}

			fmt.Println("Projects:")
			for _, project := range out.Projects {
				fmt.Printf("\t- %s\n", project.String())
			}

			return nil
		},
	}
}

func runList(cfg config.Config, orgID, tinaDir string) (listJSONOutput, error) {
	if orgID == "" {
		if cwd, err := os.Getwd(); err == nil {
			if tinacloudConfig, err := domain.TinaCloudConfigFromDisk(cwd, tinaDir); err == nil {
				orgID = tinacloudConfig.OrgID
			}
		}

		if orgID == "" {
			return listJSONOutput{}, errors.New("organization ID is required")
		}
	}

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, httpclient.ErrUnauthorized) {
			return listJSONOutput{}, err
		}
		return listJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	projects, err := requests.GetProjects(cfg.TinaCloud.IdentityEndpoint, tokens.AccessToken, orgID)
	if err != nil {
		return listJSONOutput{}, err
	}

	return listJSONOutput{
		LoggedIn: true,
		Projects: projects,
	}, nil
}
