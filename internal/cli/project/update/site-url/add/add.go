package add

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
	"tina.io/tinacloud-cli/internal/posthog"
	"tina.io/tinacloud-cli/internal/requests"
	"tina.io/tinacloud-cli/internal/utils"
)

type siteURLAddJSONOutput struct {
	LoggedIn bool     `json:"loggedIn"`
	SiteURLs []string `json:"site_urls"`
}

func Command(cfg config.Config, track posthog.Tracker) *command.Command {
	flagSet := flag.NewFlagSet("add", flag.ContinueOnError)
	var siteURLsFlag domain.UrlsFlag
	flagSet.Var(&siteURLsFlag, "url", "A URL to register against your TinaCloud project")
	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")

	return &command.Command{
		Name:        "add",
		Description: "Add site URLs",
		Flags:       flagSet,
		Run: func(args []string) error {
			out, err := runAdd(cfg, siteURLsFlag, *orgIDFlag, *projectIDFlag, *tinaDirFlag)
			if err == nil && out.LoggedIn {
				track(posthog.SiteURLAdded(siteURLsFlag))
			}
			if *jsonFlag {
				return utils.WriteJSONResult(os.Stdout, out, err)
			}
			if err != nil {
				return err
			}

			fmt.Printf("Project site URL(s): %s\n", out.SiteURLs)

			return nil
		},
	}
}

func runAdd(cfg config.Config, newSiteURLs domain.UrlsFlag, orgID, projectID, tinaDir string) (siteURLAddJSONOutput, error) {
	bothProvided := projectID != "" && orgID != ""
	bothEmpty := projectID == "" && orgID == ""

	if !bothProvided && !bothEmpty {
		return siteURLAddJSONOutput{}, errors.New("misconfigured flag usage: either provide both --org-id and --project-id, or neither to default to your tinacloud.json")
	}

	if bothEmpty {
		if cwd, err := os.Getwd(); err == nil {
			if tinacloudConfig, err := domain.TinaCloudConfigFromDisk(cwd, tinaDir); err == nil {
				projectID = tinacloudConfig.ProjectID
				orgID = tinacloudConfig.OrgID
			}
		}

		if projectID == "" {
			return siteURLAddJSONOutput{}, errors.New("project ID is missing from the config; please re-run the configure command")
		}

		if orgID == "" {
			return siteURLAddJSONOutput{}, errors.New("organization ID is missing from the config; please re-run the configure command")
		}
	}

	if len(newSiteURLs) == 0 {
		return siteURLAddJSONOutput{}, fmt.Errorf("must pass at least one site url")
	}
	newSiteURLs, err := utils.CleanSiteURLs(newSiteURLs)
	if err != nil {
		return siteURLAddJSONOutput{}, fmt.Errorf("failed to parse site urls: %w", err)
	}

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, httpclient.ErrUnauthorized) {
			return siteURLAddJSONOutput{}, httpclient.ErrUnauthorized
		}
		return siteURLAddJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	project, err := requests.GetProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, projectID)
	if err != nil {
		return siteURLAddJSONOutput{}, err
	}

	siteURLs := append(strings.Split(project.SiteURLs, ","), newSiteURLs...)

	seen := map[string]struct{}{}
	var uniqueSiteURLs []string

	for _, value := range siteURLs {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		// Dedup on the normalized origin so stored values match the cleaned new ones.
		key := value
		if normalized, err := utils.CleanSiteURL(value); err == nil {
			key = normalized
		}
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
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
		return siteURLAddJSONOutput{}, err
	}

	return siteURLAddJSONOutput{
		LoggedIn: true,
		SiteURLs: uniqueSiteURLs,
	}, nil
}
