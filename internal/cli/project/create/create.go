package create

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/posthog"
	"tina.io/tinacloud-cli/internal/requests"
)

type createJSONOutput struct {
	LoggedIn bool                  `json:"loggedIn"`
	Project  domain.CreatedProject `json:"project"`
}

func Create(args []string, cfg config.Config, track posthog.Tracker) error {
	flagSet := flag.NewFlagSet("projectcreate", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli project create [flags]")
		fmt.Fprintln(os.Stderr, "\nCreate a new project")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var projectNameFlag = flagSet.String("name", "", "Project name")
	var repoURLFlag = flagSet.String("repo-url", "", "GitHub Repo URL")
	var siteURLFlag = flagSet.String("site-url", "http://localhost:3000", "Site URL")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")
	_ = flagSet.Parse(args)

	out, err := runCreate(cfg, *orgIDFlag, *projectNameFlag, *repoURLFlag, *siteURLFlag, *tinaDirFlag)
	if err == nil && out.LoggedIn {
		track(posthog.ProjectCreated(out.Project.OrgID, out.Project.ProjectID, out.Project.Name, out.Project.RepoURL, out.Project.SiteURLs))
	}
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

	fmt.Println("Project successfully created!")
	fmt.Printf("\t- %s\n", out.Project.String())

	return nil
}

func runCreate(cfg config.Config, orgID, projectName, repoURL, siteURL, tinaDir string) (createJSONOutput, error) {
	if orgID == "" {
		if cwd, err := os.Getwd(); err == nil {
			if tinacloudConfig, err := domain.TinaCloudConfigFromDisk(cwd, tinaDir); err == nil {
				orgID = tinacloudConfig.OrgID
			}
		}

		if orgID == "" {
			return createJSONOutput{}, errors.New("organization ID is required")
		}
	}

	if projectName == "" {
		return createJSONOutput{}, errors.New("project name is required")
	}

	if repoURL == "" {
		return createJSONOutput{}, errors.New("repo url is required")
	}
	url, err := url.Parse(repoURL)
	if err != nil {
		return createJSONOutput{}, fmt.Errorf("repo url is invalid: %w", err)
	}
	if url.Scheme != "https" || url.Host != "github.com" {
		return createJSONOutput{}, errors.New("repo url is invalid: host must be https://github.com")
	}

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return createJSONOutput{LoggedIn: false}, nil
		}
		return createJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	body := domain.PostV2ProjectsBody{
		Name:    projectName,
		RepoURL: repoURL,
		SiteURL: siteURL,
	}

	resp, err := requests.CreateProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, orgID, body)
	if err != nil {
		return createJSONOutput{}, err
	}

	// If in a tinacms project, write information to <tina-dir>/tinacloud.json.
	// This is best-effort and must not block a successful return: being outside
	// a tinacms project is expected and stays silent, but real failures are
	// surfaced to stderr.
	if cwd, err := os.Getwd(); err == nil {
		err = domain.SaveTinaCloudConfig(cwd, tinaDir, orgID, resp.CreatedProject.ProjectID)
		if err != nil && !errors.Is(err, domain.ErrNotTinaProject) {
			fmt.Fprintf(os.Stderr, "warning: project created but failed to write tinacloud config: %v\n", err)
		}
	}

	return createJSONOutput{
		LoggedIn: true,
		Project:  resp.CreatedProject,
	}, nil
}
