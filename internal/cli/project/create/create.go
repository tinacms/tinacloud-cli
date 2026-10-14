package create

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/manifoldco/promptui"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/browser"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
	"tina.io/tinacloud-cli/internal/posthog"
	"tina.io/tinacloud-cli/internal/requests"
	"tina.io/tinacloud-cli/internal/utils"
)

type createJSONOutput struct {
	LoggedIn bool                  `json:"loggedIn"`
	Project  domain.CreatedProject `json:"project"`
}

func Command(cfg config.Config, track posthog.Tracker) *command.Command {
	flagSet := flag.NewFlagSet("list", flag.ContinueOnError)
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var projectNameFlag = flagSet.String("name", "", "Project name")
	var repoURLFlag = flagSet.String("repo-url", "", "GitHub Repo URL")
	var siteURLFlag = flagSet.String("site-url", "http://localhost:3000", "Site URL")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")
	var noInputFlag = flagSet.Bool("no-input", false, "Fail instead of prompting for input")

	return &command.Command{
		Name:          "create",
		Description:   "Create a new project",
		Example:       "tinacloud project create --name my-project --repo-url https://github.com/acme/my-project",
		RequiresInput: true,
		Flags:         flagSet,
		Run: func(args []string) error {
			out, err := runCreate(cfg, browser.OpenURL, !*noInputFlag && utils.IsTerminal(os.Stdin), *orgIDFlag, *projectNameFlag, *repoURLFlag, *siteURLFlag, *tinaDirFlag)
			if err == nil && out.LoggedIn {
				track(posthog.ProjectCreated(out.Project.OrgID, out.Project.ProjectID, out.Project.Name, out.Project.RepoURL, out.Project.SiteURLs))
			}
			if *jsonFlag {
				return utils.WriteJSONResult(out, err)
			}
			if err != nil {
				return err
			}

			fmt.Println("✅️ Successfully created project!")
			fmt.Printf("\t- %s\n", out.Project.String())

			return nil
		},
	}
}

func runCreate(cfg config.Config, openURL func(string) error, allowPrompt bool, orgID, projectName, repoURL, siteURL, tinaDir string) (createJSONOutput, error) {
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
	repoUrl, err := url.Parse(repoURL)
	if err != nil {
		return createJSONOutput{}, fmt.Errorf("repo url is invalid: %w", err)
	}
	if repoUrl.Scheme != "https" || repoUrl.Host != "github.com" {
		return createJSONOutput{}, errors.New("repo url is invalid: host must be https://github.com")
	}

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, httpclient.ErrUnauthorized) {
			return createJSONOutput{}, err
		}
		return createJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	body := domain.PostV2ProjectsBody{
		Name:    projectName,
		RepoURL: repoURL,
		SiteURL: siteURL,
	}

	var resp domain.PostV2ProjectResponse
	resp, err = requests.CreateProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, orgID, body)
	if err != nil {
		if !errors.Is(err, requests.ErrMissingInstallationID) {
			return createJSONOutput{}, err
		}
		if !allowPrompt {
			return createJSONOutput{}, errors.New("GitHub App installation is required; rerun interactively to complete installation")
		}

		ln, port, err := utils.CallbackListener()
		if err != nil {
			return createJSONOutput{}, err
		}
		defer ln.Close()

		{
			githubGitHost := domain.GitHost{ID: "github", Name: "GitHub", Server: "github.com"}
			githosts := []domain.GitHost{githubGitHost}

			githostResp, err := requests.GetGitHosts(cfg.TinaCloud.Endpoint, tokens.AccessToken, orgID)
			if err != nil {
				return createJSONOutput{}, fmt.Errorf("failed to retrieve githosts: %w", err)
			}
			githosts = append(githosts, githostResp.GitHosts...)

			state := fmt.Sprintf("http: localhost:%d", port)
			installationURL, err := githubAppInstallationURL(cfg, githubGitHost, state)
			if err != nil {
				return createJSONOutput{}, fmt.Errorf("failed to create github installation url: %w", err)
			}
			if len(githosts) > 1 {
				var menuItems []string
				for _, githost := range githosts {
					menuItems = append(
						menuItems,
						fmt.Sprintf("%s - %s", githost.Name, githost.Server),
					)
				}

				prompt := promptui.Select{
					Label: "Please select the GitHost you would like to install the app into",
					Items: menuItems,
					Size:  10,
				}
				selectedIndex, _, err := prompt.Run()
				if err != nil {
					return createJSONOutput{}, fmt.Errorf("failed to select githost: %w", err)
				}

				githost := githosts[selectedIndex]
				installationURL, err = githubAppInstallationURL(cfg, githost, state)
				if err != nil {
					return createJSONOutput{}, fmt.Errorf("failed to create github installation url: %w", err)
				}
			}

			// User can copy string from terminal if this fails
			_ = openURL(installationURL)

			fmt.Fprintf(os.Stderr, `Install the TinaCloud GitHub App.
Attempting to open your default browser. If the browser does not open, open the following URL.

%s

`, installationURL)
		}

		var installationID *int64
		{
			conn, err := ln.Accept()
			if err != nil {
				return createJSONOutput{}, fmt.Errorf("failed to accept auth callback: %w", err)
			}
			defer conn.Close()

			req, err := http.ReadRequest(bufio.NewReader(conn))
			if err != nil {
				return createJSONOutput{}, fmt.Errorf("failed to read auth callback: %w", err)
			}

			err = utils.CallbackResponse(conn)
			if err != nil {
				return createJSONOutput{}, err
			}

			params, err := url.ParseQuery(req.URL.RawQuery)
			if err != nil {
				return createJSONOutput{}, fmt.Errorf("failed to read installation_id from auth callback: %w", err)
			}

			installationIDParam := params.Get("installation_id")
			if installationIDParam == "" {
				return createJSONOutput{}, errors.New("installation_id is missing from client request")
			}

			id, err := strconv.ParseInt(installationIDParam, 10, 64)
			if err != nil {
				return createJSONOutput{}, fmt.Errorf("failed to parse installation_id from auth callback: %w", err)
			}

			installationID = &id
		}

		body.InstallationID = installationID
		resp, err = requests.CreateProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, orgID, body)
		if err != nil {
			return createJSONOutput{}, err
		}
	}

	// Attempt to write information to <tina-dir>/tinacloud.json if in a
	// tinacms project This is best-effort and must not block a successful
	// return
	if cwd, err := os.Getwd(); err == nil {
		err = domain.SaveTinaCloudConfig(cwd, tinaDir, orgID, resp.CreatedProject.ProjectID)
		if err != nil && !errors.Is(err, domain.ErrNotTinaProject) {
			fmt.Fprintf(os.Stderr, "failed to write tinacloud config: %v\n", err)
		}
	}

	return createJSONOutput{
		LoggedIn: true,
		Project:  resp.CreatedProject,
	}, nil
}

func githubAppInstallationURL(cfg config.Config, githost domain.GitHost, state string) (string, error) {
	if githost.Server != "" && githost.GitHubAppName != "" {
		return fmt.Sprintf("https://%s/github-apps/%s/installations/new?state=%s", githost.Server, githost.GitHubAppName, state), nil
	}

	githubApp, err := requests.GetTinaCloudGitHubApp(cfg.TinaCloud.Endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve tinacloud github app name: %w", err)
	}

	return fmt.Sprintf("https://%s/apps/%s/installations/new?state=%s", githost.Server, githubApp.Name, state), nil
}
