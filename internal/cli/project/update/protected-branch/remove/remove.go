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
	"slices"
	"strings"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
)

type protectedBranchRemoveJSONOutput struct {
	LoggedIn          bool     `json:"loggedIn"`
	ProtectedBranches []string `json:"protected_branches"`
}

func Remove(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("projectupdateprotectedbranchremove", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli project update protected-branch remove [flags]")
		fmt.Fprintln(os.Stderr, "\nRemove protected branches")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var protectedBranchesFlag domain.BranchesFlag
	flagSet.Var(&protectedBranchesFlag, "branch", "A protected branch to remove from your TinaCloud project")
	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	_ = flagSet.Parse(args)

	out, err := runRemove(cfg, protectedBranchesFlag, *orgIDFlag, *projectIDFlag)
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

	fmt.Printf("Project protected branch(es): %s\n", out.ProtectedBranches)

	return nil
}

func runRemove(cfg config.Config, removeProtectedBranches []string, orgID, projectID string) (protectedBranchRemoveJSONOutput, error) {
	if projectID == "" {
		return protectedBranchRemoveJSONOutput{}, errors.New("project ID is required")
	}

	if orgID == "" {
		return protectedBranchRemoveJSONOutput{}, errors.New("organization ID is required")
	}

	if len(removeProtectedBranches) == 0 {
		return protectedBranchRemoveJSONOutput{}, errors.New("must pass at least one protected branch")
	}

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return protectedBranchRemoveJSONOutput{LoggedIn: false}, nil
		}
		return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	client := http.Client{}

	var project domain.GetV2ProjectResponse
	{
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/apps/%s", cfg.TinaCloud.Endpoint, projectID), nil)
		if err != nil {
			return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to get projects: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return protectedBranchRemoveJSONOutput{}, auth.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to read get projects response: %w", err)
		}

		if resp.StatusCode > 299 {
			return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to get projects. received response status of '%d': %s", resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &project); err != nil {
			return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to unmarshal project from get request: %w", err)
		}
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

	{
		putBody, err := json.Marshal(domain.PutV2ProjectBody{
			Name:              project.Name,
			RepoURL:           project.RepoURL,
			InstallationID:    project.InstallationID,
			PathToTina:        project.PathToTina,
			Features:          project.Features,
			SiteURL:           project.SiteURLs,
			ProtectedBranches: &protectedBranches,
		})
		if err != nil {
			return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to marshal request body: %w", err)
		}

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v2/organizations/%s/apps/%s", cfg.TinaCloud.Endpoint, orgID, projectID), bytes.NewBuffer(putBody))
		if err != nil {
			return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to update project: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return protectedBranchRemoveJSONOutput{}, auth.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to read update project response: %w", err)
		}

		if resp.StatusCode > 299 {
			return protectedBranchRemoveJSONOutput{}, fmt.Errorf("failed to update project. received response status of '%d': %s", resp.StatusCode, string(body))
		}
	}

	return protectedBranchRemoveJSONOutput{
		LoggedIn:          true,
		ProtectedBranches: protectedBranches,
	}, nil
}
