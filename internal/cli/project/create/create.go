package create

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
)

type createJSONOutput struct {
	LoggedIn bool                  `json:"loggedIn"`
	Project  domain.CreatedProject `json:"project"`
}

func Create(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("projectcreate", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli project create [flags]")
		fmt.Fprintln(os.Stderr, "\nCreate a new project")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var nameFlag = flagSet.String("name", "", "Project name")
	var repoURLFlag = flagSet.String("repo-url", "", "GitHub Repo URL")
	var siteURLFlag = flagSet.String("site-url", "http://localhost:3000", "Site URL")
	_ = flagSet.Parse(args)

	out, err := runCreate(cfg, *orgIDFlag, *nameFlag, *repoURLFlag, *siteURLFlag)
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

func runCreate(cfg config.Config, orgIDFlag, nameFlag, repoURLFlag, siteURLFlag string) (createJSONOutput, error) {
	if orgIDFlag == "" {
		return createJSONOutput{}, errors.New("organization ID is required")
	}

	if nameFlag == "" {
		return createJSONOutput{}, errors.New("project name is required")
	}

	if repoURLFlag == "" {
		return createJSONOutput{}, errors.New("repo url is required")
	}
	url, err := url.Parse(repoURLFlag)
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

	createProjectBody := domain.PostV2ProjectsBody{
		Name:    nameFlag,
		RepoURL: repoURLFlag,
		SiteURL: siteURLFlag,
	}

	postBody, err := json.Marshal(createProjectBody)
	if err != nil {
		return createJSONOutput{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v2/organizations/%s/apps", cfg.TinaCloud.Endpoint, orgIDFlag), bytes.NewBuffer(postBody))
	if err != nil {
		return createJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return createJSONOutput{}, fmt.Errorf("failed to create project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return createJSONOutput{}, auth.ErrUnauthorized
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return createJSONOutput{}, fmt.Errorf("failed to read create project response: %w", err)
	}

	if resp.StatusCode > 299 {
		return createJSONOutput{}, fmt.Errorf("failed to create project. received response status of '%d': %s", resp.StatusCode, string(responseBody))
	}

	var projectCreateResponse domain.PostV2ProjectResponse
	if err := json.Unmarshal(responseBody, &projectCreateResponse); err != nil {
		return createJSONOutput{}, fmt.Errorf("failed to unmarshal created project from post request: %w", err)
	}

	return createJSONOutput{
		LoggedIn: true,
		Project:  projectCreateResponse.CreatedProject,
	}, nil
}
