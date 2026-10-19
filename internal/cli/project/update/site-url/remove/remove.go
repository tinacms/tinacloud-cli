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
	"tina.io/tinacloud-cli/internal/httpclient"
	"tina.io/tinacloud-cli/internal/requests"
	"tina.io/tinacloud-cli/internal/utils"
)

type siteURLRemoveJSONOutput struct {
	LoggedIn bool     `json:"loggedIn"`
	SiteURLs []string `json:"site_urls"`
}

func Command(cfg config.Config) *command.Command {
	flagSet := flag.NewFlagSet("remove", flag.ContinueOnError)
	var siteURLsFlag domain.UrlsFlag
	flagSet.Var(&siteURLsFlag, "url", "A URL to remove from your TinaCloud project")
	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")

	return &command.Command{
		Name:          "remove",
		Description:   "Remove site URLs",
		Example:       "tinacloud project update site-url remove --url https://example.com",
		RequiresInput: true,
		Flags:         flagSet,
		Run: func(args []string) error {
			out, err := runRemove(cfg, siteURLsFlag, *orgIDFlag, *projectIDFlag, *tinaDirFlag)
			if *jsonFlag {
				return utils.WriteJSONResult(out, err)
			}
			if err != nil {
				return err
			}

			fmt.Printf("Project site URL(s): %s\n", out.SiteURLs)

			return nil
		},
	}
}

func runRemove(cfg config.Config, removeSiteURLs domain.UrlsFlag, orgID, projectID, tinaDir string) (siteURLRemoveJSONOutput, error) {
	bothProvided := projectID != "" && orgID != ""
	bothEmpty := projectID == "" && orgID == ""

	if !bothProvided && !bothEmpty {
		return siteURLRemoveJSONOutput{}, errors.New("misconfigured flag usage: either provide both --org-id and --project-id, or neither to default to your tinacloud.json")
	}

	if bothEmpty {
		if cwd, err := os.Getwd(); err == nil {
			if tinacloudConfig, err := domain.TinaCloudConfigFromDisk(cwd, tinaDir); err == nil {
				projectID = tinacloudConfig.ProjectID
				orgID = tinacloudConfig.OrgID
			}
		}

		if projectID == "" {
			return siteURLRemoveJSONOutput{}, errors.New("project ID is missing from the config; please re-run the configure command")
		}

		if orgID == "" {
			return siteURLRemoveJSONOutput{}, errors.New("organization ID is missing from the config; please re-run the configure command")
		}
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
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, httpclient.ErrUnauthorized) {
			return siteURLRemoveJSONOutput{}, err
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

		// Match on the normalized origin so trailing-slash/path variants still remove.
		key := value
		if normalized, err := utils.CleanSiteURL(value); err == nil {
			key = normalized
		}
		if _, exists := removeSet[key]; exists {
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
