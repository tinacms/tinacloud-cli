package remove

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

type protectedBranchRemoveJSONOutput struct {
	LoggedIn          bool     `json:"loggedIn"`
	ProtectedBranches []string `json:"protected_branches"`
}

func Command(cfg config.Config, track posthog.Tracker) *command.Command {
	flagSet := flag.NewFlagSet("remove", flag.ContinueOnError)
	var protectedBranchesFlag domain.BranchesFlag
	flagSet.Var(&protectedBranchesFlag, "branch", "A protected branch to remove from your TinaCloud project")
	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")

	return &command.Command{
		Name:        "remove",
		Description: "Remove protected branches",
		Flags:       flagSet,
		Run: func(args []string) error {
			out, err := runRemove(cfg, protectedBranchesFlag, *orgIDFlag, *projectIDFlag, *tinaDirFlag)
			if err == nil && out.LoggedIn {
				track(posthog.ProtectedBranchRemoved(protectedBranchesFlag))
			}
			if *jsonFlag {
				return utils.WriteJSONResult(os.Stdout, out, err)
			}
			if err != nil {
				return err
			}

			fmt.Printf("Project protected branch(es): %s\n", out.ProtectedBranches)

			return nil
		},
	}
}

func runRemove(cfg config.Config, removeProtectedBranches []string, orgID, projectID, tinaDir string) (protectedBranchRemoveJSONOutput, error) {
	bothProvided := projectID != "" && orgID != ""
	bothEmpty := projectID == "" && orgID == ""

	if !bothProvided && !bothEmpty {
		return protectedBranchRemoveJSONOutput{}, errors.New("misconfigured flag usage: either provide both --org-id and --project-id, or neither to default to your tinacloud.json")
	}

	if bothEmpty {
		if cwd, err := os.Getwd(); err == nil {
			if tinacloudConfig, err := domain.TinaCloudConfigFromDisk(cwd, tinaDir); err == nil {
				projectID = tinacloudConfig.ProjectID
				orgID = tinacloudConfig.OrgID
			}
		}

		if projectID == "" {
			return protectedBranchRemoveJSONOutput{}, errors.New("project ID is missing from the config; please re-run the configure command")
		}

		if orgID == "" {
			return protectedBranchRemoveJSONOutput{}, errors.New("organization ID is missing from the config; please re-run the configure command")
		}
	}

	if len(removeProtectedBranches) == 0 {
		return protectedBranchRemoveJSONOutput{}, errors.New("must pass at least one protected branch")
	}

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, httpclient.ErrUnauthorized) {
			return protectedBranchRemoveJSONOutput{}, httpclient.ErrUnauthorized
		}
		return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	project, err := requests.GetProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, projectID)
	if err != nil {
		return protectedBranchRemoveJSONOutput{}, err
	}

	if project.Features == nil || !slices.Contains(*project.Features, "editorial-workflow") {
		return protectedBranchRemoveJSONOutput{}, errors.New("protected branches are a feature of the editorial workflow, please upgrade your plan to get access")
	}

	if project.ProtectedBranches == nil {
		return protectedBranchRemoveJSONOutput{
			LoggedIn:          true,
			ProtectedBranches: []string{},
		}, nil
	}

	removeSet := map[string]struct{}{}
	for _, value := range removeProtectedBranches {
		removeSet[strings.TrimSpace(value)] = struct{}{}
	}

	var protectedBranches []string
	for _, value := range *project.ProtectedBranches {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := removeSet[value]; exists {
			continue
		}

		protectedBranches = append(protectedBranches, value)
	}

	body := domain.PutV2ProjectBody{
		Name:              project.Name,
		RepoURL:           project.RepoURL,
		InstallationID:    project.InstallationID,
		PathToTina:        project.PathToTina,
		Features:          project.Features,
		SiteURL:           project.SiteURLs,
		ProtectedBranches: &protectedBranches,
	}

	err = requests.PutProject(cfg.TinaCloud.Endpoint, tokens.AccessToken, orgID, projectID, body)
	if err != nil {
		return protectedBranchRemoveJSONOutput{}, err
	}

	return protectedBranchRemoveJSONOutput{
		LoggedIn:          true,
		ProtectedBranches: protectedBranches,
	}, nil
}
