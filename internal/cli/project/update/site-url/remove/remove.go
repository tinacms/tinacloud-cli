package remove

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

type siteURLRemoveJSONOutput struct {
	LoggedIn bool     `json:"loggedIn"`
	SiteURLs []string `json:"site_urls"`
}

func Remove(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("projectupdatesiteurlremove", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli project update site-url remove [flags]")
		fmt.Fprintln(os.Stderr, "\nRemove site URLs")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var siteURLsFlag domain.UrlsFlag
	flagSet.Var(&siteURLsFlag, "url", "A URL to remove from your TinaCloud project")
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

	out, err := runRemove(cfg, siteURLs, *orgIDFlag, *projectIDFlag)
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

	fmt.Printf("Removed site URL(s): %s\n", siteURLs)
	fmt.Printf("\t- Project site URL(s): %s\n", out.SiteURLs)

	return nil
}

func runRemove(cfg config.Config, removeSiteURLs []string, orgIDFlag, projectIDFlag string) (siteURLRemoveJSONOutput, error) {
	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return siteURLRemoveJSONOutput{LoggedIn: false}, nil
		}
		return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	client := http.Client{}

	var project domain.GetV2ProjectResponse
	{
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/apps/%s", cfg.TinaCloud.Endpoint, projectIDFlag), nil)
		if err != nil {
			return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to get projects: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return siteURLRemoveJSONOutput{}, auth.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to read get projects response: %w", err)
		}

		if resp.StatusCode > 299 {
			return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to get projects. received response status of '%d': %s", resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &project); err != nil {
			return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to unmarshal project from get request: %w", err)
		}
	}

	removeSet := map[string]struct{}{}
	for _, value := range removeSiteURLs {
		removeSet[value] = struct{}{}
	}

	var siteURLs []string
	for _, value := range strings.Split(project.SiteURLs, ",") {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if _, exists := removeSet[value]; exists {
			continue
		}

		siteURLs = append(siteURLs, value)
	}

	{
		putBody, err := json.Marshal(domain.PutV2ProjectBody{
			Name:              project.Name,
			RepoURL:           project.RepoURL,
			InstallationID:    project.InstallationID,
			PathToTina:        project.PathToTina,
			ProtectedBranches: project.ProtectedBranches,
			Features:          project.Features,
			SiteURL:           strings.Join(siteURLs, ","),
		})
		if err != nil {
			return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to marshal request body: %w", err)
		}

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v2/organizations/%s/apps/%s", cfg.TinaCloud.Endpoint, orgIDFlag, projectIDFlag), bytes.NewBuffer(putBody))
		if err != nil {
			return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("ContentType", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to update project: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return siteURLRemoveJSONOutput{}, auth.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to read update project response: %w", err)
		}

		if resp.StatusCode > 299 {
			return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to update project. received response status of '%d': %s", resp.StatusCode, string(body))
		}
	}

	return siteURLRemoveJSONOutput{
		LoggedIn: true,
		SiteURLs: siteURLs,
	}, nil
}
