package remove

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/requests"
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

	out, err := runRemove(cfg, siteURLsFlag, *orgIDFlag, *projectIDFlag)
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

	fmt.Printf("Project site URL(s): %s\n", out.SiteURLs)

	return nil
}

func runRemove(cfg config.Config, removeSiteURLs domain.UrlsFlag, orgID, projectID string) (siteURLRemoveJSONOutput, error) {
	if projectID == "" {
		return siteURLRemoveJSONOutput{}, fmt.Errorf("project ID is required")
	}

	if orgID == "" {
		return siteURLRemoveJSONOutput{}, fmt.Errorf("organization ID is required")
	}

	if len(removeSiteURLs) == 0 {
		return siteURLRemoveJSONOutput{}, fmt.Errorf("must pass at least one site url")
	}
	removeSiteURLs, err := utils.CleanSiteURLs(removeSiteURLs)
	if err != nil {
		return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to parse site urls: %w", err)
	}

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return siteURLRemoveJSONOutput{LoggedIn: false}, nil
		}
		return siteURLRemoveJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	project, err := requests.GetProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, projectID)
	if err != nil {
		return siteURLRemoveJSONOutput{}, err
	}

	removeSet := map[string]struct{}{}
	for _, value := range removeSiteURLs {
		removeSet[value] = struct{}{}
	}

	var uniqueSiteURLs []string
	for _, value := range strings.Split(project.SiteURLs, ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := removeSet[value]; exists {
			continue
		}

		uniqueSiteURLs = append(uniqueSiteURLs, value)
	}

	body := domain.PutV2ProjectBody{
		Name:              project.Name,
		RepoURL:           project.RepoURL,
		InstallationID:    project.InstallationID,
		PathToTina:        project.PathToTina,
		ProtectedBranches: project.ProtectedBranches,
		Features:          project.Features,
		SiteURL:           strings.Join(uniqueSiteURLs, ","),
	}

	err = requests.PutProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, orgID, projectID, body)
	if err != nil {
		return siteURLRemoveJSONOutput{}, err
	}

	return siteURLRemoveJSONOutput{
		LoggedIn: true,
		SiteURLs: uniqueSiteURLs,
	}, nil
}
