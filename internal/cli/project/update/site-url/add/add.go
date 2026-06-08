package add

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/utils"
)

type siteURLAddJSONOutput struct {
	LoggedIn bool     `json:"loggedIn"`
	SiteURLs []string `json:"site_urls"`
}

func Add(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("projectupdatesiteurladd", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli project update site-url add [flags]")
		fmt.Fprintln(os.Stderr, "\nAdd site URLs")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var siteURLsFlag domain.UrlsFlag
	flagSet.Var(&siteURLsFlag, "url", "A URL to register against your TinaCloud project")
	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	_ = flagSet.Parse(args)

	if *projectIDFlag == "" {
		return fmt.Errorf("project ID is required")
	}

	if *orgIDFlag == "" {
		return fmt.Errorf("organization ID is required")
	}

	if len(siteURLsFlag) == 0 {
		return fmt.Errorf("must pass at least one site url")
	}
	siteURLs, err := utils.CleanSiteURLs(siteURLsFlag)
	if err != nil {
		return fmt.Errorf("failed to parse site urls: %w", err)
	}

	out, err := runAdd(cfg, siteURLs, *orgIDFlag, *projectIDFlag)
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

	fmt.Printf("Added site URL(s): %s\n", siteURLs)
	fmt.Printf("\t- Project site URL(s): %s\n", out.SiteURLs)

	return nil
}

func runAdd(cfg config.Config, newSiteURLs []string, orgIDFlag, projectIDFlag string) (siteURLAddJSONOutput, error) {
	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return siteURLAddJSONOutput{LoggedIn: false}, nil
		}
		return siteURLAddJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	client := http.Client{}

	var project domain.GetV2ProjectResponse
	{
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/apps/%s", cfg.TinaCloud.Endpoint, projectIDFlag), nil)
		if err != nil {
			return siteURLAddJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return siteURLAddJSONOutput{}, fmt.Errorf("failed to get projects: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return siteURLAddJSONOutput{}, auth.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return siteURLAddJSONOutput{}, fmt.Errorf("failed to read get projects response: %w", err)
		}

		if resp.StatusCode > 299 {
			return siteURLAddJSONOutput{}, fmt.Errorf("failed to get projects. received response status of '%d': %s", resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &project); err != nil {
			return siteURLAddJSONOutput{}, fmt.Errorf("failed to unmarshal project from get request: %w", err)
		}
	}

	var siteURLs []string
	for _, url := range strings.Split(project.SitURLs, ",") {
		siteURLs = append(siteURLs, url)
	}
	for _, url := range newSiteURLs {
		siteURLs = append(siteURLs, url)
	}

	seen := map[string]struct{}{}
	var uniqueSiteURLs []string

	for _, value := range siteURLs {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}
		uniqueSiteURLs = append(uniqueSiteURLs, value)
	}

	{
		putBody, err := json.Marshal(domain.PutV2ProjectBody{
			Name:              project.Name,
			RepoURL:           project.RepoURL,
			InstallationID:    project.InstallationID,
			PathToTina:        project.PathToTina,
			ProtectedBranches: project.ProtectedBranches,
			Features:          project.Features,
			SitURL:            strings.Join(uniqueSiteURLs, ","),
		})
		if err != nil {
			return siteURLAddJSONOutput{}, fmt.Errorf("failed to marshal request body: %w", err)
		}

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v2/organizations/%s/apps/%s", cfg.TinaCloud.Endpoint, orgIDFlag, projectIDFlag), bytes.NewBuffer(putBody))
		if err != nil {
			return siteURLAddJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("ContentType", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return siteURLAddJSONOutput{}, fmt.Errorf("failed to get projects: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return siteURLAddJSONOutput{}, auth.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return siteURLAddJSONOutput{}, fmt.Errorf("failed to read get projects response: %w", err)
		}

		if resp.StatusCode > 299 {
			return siteURLAddJSONOutput{}, fmt.Errorf("failed to get projects. received response status of '%d': %s", resp.StatusCode, string(body))
		}
	}

	return siteURLAddJSONOutput{
		LoggedIn: true,
		SiteURLs: uniqueSiteURLs,
	}, nil
}
