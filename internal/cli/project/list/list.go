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
	LoggedIn bool                           `json:"loggedIn"`
	Projects []domain.GetV2ProjectsResponse `json:"projects"`
}

func List(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("projectlist", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud project list [flags]")
		fmt.Fprintln(os.Stderr, "\nList projects")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")
	_ = flagSet.Parse(args)

	out, err := runList(cfg, *orgIDFlag, *tinaDirFlag)
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

	fmt.Println("Projects:")
	for _, project := range out.Projects {
		fmt.Printf("\t- %s\n", project.String())
	}

	return nil
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
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return listJSONOutput{LoggedIn: false}, nil
		}
		return listJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	projects, err := requests.GetProjects(cfg.TinaCloud.Endpoint, tokens.AccessToken, orgID)
	if err != nil {
		return listJSONOutput{}, err
	}

	return listJSONOutput{
		LoggedIn: true,
		Projects: projects,
	}, nil
}
