package add

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
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

type protectedBranchAddJSONOutput struct {
	LoggedIn          bool     `json:"loggedIn"`
	ProtectedBranches []string `json:"protected_branches"`
}

func Command(cfg config.Config, track posthog.Tracker) *command.Command {
	flagSet := flag.NewFlagSet("add", flag.ContinueOnError)
	var protectedBranchesFlag domain.BranchesFlag
	flagSet.Var(&protectedBranchesFlag, "branch", "A protected branch to add to your TinaCloud project")
	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")

	return &command.Command{
		Name:        "add",
		Description: "Add protected branches",
		Flags:       flagSet,
		Run: func(args []string) error {
			out, err := runAdd(cfg, protectedBranchesFlag, *orgIDFlag, *projectIDFlag, *tinaDirFlag)
			if err == nil && out.LoggedIn {
				track(posthog.ProtectedBranchAdded(protectedBranchesFlag))
			}
			if *jsonFlag {
				return utils.WriteJSONResult(os.Stdout, out, err)
			}
			if err != nil {
				return err
			}

			if !out.LoggedIn {
				fmt.Println("You are not currently logged in")
				return nil
			}

			fmt.Printf("Project protected branch(es): %s\n", out.ProtectedBranches)

			return nil
		},
	}
}

func runAdd(cfg config.Config, newProtectedBranches []string, orgID, projectID, tinaDir string) (protectedBranchAddJSONOutput, error) {
	bothProvided := projectID != "" && orgID != ""
	bothEmpty := projectID == "" && orgID == ""

	if !bothProvided && !bothEmpty {
		return protectedBranchAddJSONOutput{}, errors.New("misconfigured flag usage: either provide both --org-id and --project-id, or neither to default to your tinacloud.json")
	}

	if bothEmpty {
		if cwd, err := os.Getwd(); err == nil {
			if tinacloudConfig, err := domain.TinaCloudConfigFromDisk(cwd, tinaDir); err == nil {
				projectID = tinacloudConfig.ProjectID
				orgID = tinacloudConfig.OrgID
			}
		}

		if projectID == "" {
			return protectedBranchAddJSONOutput{}, errors.New("project ID is missing from the config; please re-run the configure command")
		}

		if orgID == "" {
			return protectedBranchAddJSONOutput{}, errors.New("organization ID is missing from the config; please re-run the configure command")
		}
	}

	if len(newProtectedBranches) == 0 {
		return protectedBranchAddJSONOutput{}, errors.New("must pass at least one protected branch")
	}

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, httpclient.ErrUnauthorized) {
			return protectedBranchAddJSONOutput{LoggedIn: false}, nil
		}
		return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	project, err := requests.GetProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, projectID)
	if err != nil {
		return protectedBranchAddJSONOutput{}, err
	}

	if project.Features == nil || !slices.Contains(*project.Features, "editorial-workflow") {
		return protectedBranchAddJSONOutput{}, errors.New("protected branches are a feature of the editorial workflow, please upgrade your plan to get access")
	}

	var protectedBranches []string
	if project.ProtectedBranches != nil {
		protectedBranches = append(protectedBranches, (*project.ProtectedBranches)...)
	}
	protectedBranches = append(protectedBranches, newProtectedBranches...)

	seen := map[string]struct{}{}
	var uniqueProtectedBranches []string

	for _, value := range protectedBranches {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}
		uniqueProtectedBranches = append(uniqueProtectedBranches, value)
	}

	body := domain.PutV2ProjectBody{
		Name:              project.Name,
		RepoURL:           project.RepoURL,
		InstallationID:    project.InstallationID,
		PathToTina:        project.PathToTina,
		Features:          project.Features,
		SiteURL:           project.SiteURLs,
		ProtectedBranches: &uniqueProtectedBranches,
	}

	err = requests.PutProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, orgID, projectID, body)
	if err != nil {
		return protectedBranchAddJSONOutput{}, err
	}

	return protectedBranchAddJSONOutput{
		LoggedIn:          true,
		ProtectedBranches: uniqueProtectedBranches,
	}, nil
}
