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
	"slices"
	"strings"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/posthog"
)

type protectedBranchAddJSONOutput struct {
	LoggedIn          bool     `json:"loggedIn"`
	ProtectedBranches []string `json:"protected_branches"`
}

func Add(args []string, cfg config.Config, track posthog.Tracker) error {
	flagSet := flag.NewFlagSet("projectupdateprotectedbranchadd", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli project update protected-branch add [flags]")
		fmt.Fprintln(os.Stderr, "\nAdd protected branches")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var protectedBranchesFlag domain.BranchesFlag
	flagSet.Var(&protectedBranchesFlag, "branch", "A protected branch to add to your TinaCloud project")
	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	_ = flagSet.Parse(args)

	out, err := runAdd(cfg, protectedBranchesFlag, *orgIDFlag, *projectIDFlag)
	if err == nil && out.LoggedIn {
		track(posthog.ProtectedBranchAdded(protectedBranchesFlag))
	}
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

func runAdd(cfg config.Config, newProtectedBranches []string, orgID, projectID string) (protectedBranchAddJSONOutput, error) {
	if projectID == "" {
		return protectedBranchAddJSONOutput{}, errors.New("project ID is required")
	}

	if orgID == "" {
		return protectedBranchAddJSONOutput{}, errors.New("organization ID is required")
	}

	if len(newProtectedBranches) == 0 {
		return protectedBranchAddJSONOutput{}, errors.New("must pass at least one protected branch")
	}

	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return protectedBranchAddJSONOutput{LoggedIn: false}, nil
		}
		return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	client := http.Client{}

	var project domain.GetV2ProjectResponse
	{
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/apps/%s", cfg.TinaCloud.Endpoint, projectID), nil)
		if err != nil {
			return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to get projects: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return protectedBranchAddJSONOutput{}, auth.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to read get projects response: %w", err)
		}

		if resp.StatusCode > 299 {
			return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to get projects. received response status of '%d': %s", resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &project); err != nil {
			return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to unmarshal project from get request: %w", err)
		}
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

	{
		putBody, err := json.Marshal(domain.PutV2ProjectBody{
			Name:              project.Name,
			RepoURL:           project.RepoURL,
			InstallationID:    project.InstallationID,
			PathToTina:        project.PathToTina,
			Features:          project.Features,
			SiteURL:           project.SiteURLs,
			ProtectedBranches: &uniqueProtectedBranches,
		})
		if err != nil {
			return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to marshal request body: %w", err)
		}

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v2/organizations/%s/apps/%s", cfg.TinaCloud.Endpoint, orgID, projectID), bytes.NewBuffer(putBody))
		if err != nil {
			return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to update project: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return protectedBranchAddJSONOutput{}, auth.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to read update project response: %w", err)
		}

		if resp.StatusCode > 299 {
			return protectedBranchAddJSONOutput{}, fmt.Errorf("failed to update project. received response status of '%d': %s", resp.StatusCode, string(body))
		}
	}

	return protectedBranchAddJSONOutput{
		LoggedIn:          true,
		ProtectedBranches: uniqueProtectedBranches,
	}, nil
}
