package configure

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/chzyer/readline"
	"github.com/manifoldco/promptui"
	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
)

type projectOption struct {
	project domain.GetV2ProjectsResponse
	org     domain.Organization
}

func Configure(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("configure", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli configure [flags]")
		fmt.Fprintln(os.Stderr, "\nConfigure a local Tina project")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")
	_ = flagSet.Parse(args)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	saved, err := runConfigure(cfg, cwd, *orgIDFlag, *projectIDFlag, *tinaDirFlag, auth.TokensFromCache)
	if err != nil {
		return err
	}

	if saved {
		fmt.Printf("Wrote tinacloud config to %s\n", path.Join(cwd, *tinaDirFlag, domain.TinaCloudConfigFileName))
	} else {
		fmt.Println("Did not write to tinacloud config")
	}

	return nil
}

func runConfigure(cfg config.Config, cwd, orgID, projectID, tinaDir string, loadTokens auth.TokenLoader) (bool, error) {
	if orgID != "" || projectID != "" {
		if orgID == "" || projectID == "" {
			return false, fmt.Errorf("both --org-id and --project-id must be passed")
		}
		return true, domain.SaveTinaCloudConfig(cwd, tinaDir, orgID, projectID)
	}

	tokens, err := loadTokens(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return false, nil
		}
		return false, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	client := http.Client{}

	var organizations []domain.OrganizationWithRole
	{
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/organizations", cfg.TinaCloud.Endpoint), nil)
		if err != nil {
			return false, fmt.Errorf("failed to initialize request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return false, fmt.Errorf("failed to get organizations: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return false, auth.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("failed to read get organizations response: %w", err)
		}

		if resp.StatusCode > 299 {
			return false, fmt.Errorf("failed to get organizations. received response status of '%d': %s", resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &organizations); err != nil {
			return false, fmt.Errorf("failed to unmarshal organizations from get request: %w", err)
		}
	}

	var projectOptions []projectOption
	for _, org := range organizations {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/organizations/%s/apps", cfg.TinaCloud.Endpoint, org.Organization.ID), nil)
		if err != nil {
			return false, fmt.Errorf("failed to initialize request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return false, fmt.Errorf("failed to get projects: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return false, auth.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("failed to read get projects response: %w", err)
		}

		if resp.StatusCode > 299 {
			return false, fmt.Errorf("failed to get projects. received response status of '%d': %s", resp.StatusCode, string(body))
		}

		var projectsResponse []domain.GetV2ProjectsResponse
		if err := json.Unmarshal(body, &projectsResponse); err != nil {
			return false, fmt.Errorf("failed to unmarshal projects from get request: %w", err)
		}

		for _, project := range projectsResponse {
			projectOptions = append(
				projectOptions,
				projectOption{
					project: project,
					org:     org.Organization,
				},
			)
		}
	}

	if len(organizations) == 1 && len(projectOptions) == 1 {
		org := projectOptions[0].org
		project := projectOptions[0].project

		confirmed, err := confirmPrompt(fmt.Sprintf("Continue with %s (%s)?", project.Name, org.Name))
		if err != nil {
			return false, fmt.Errorf("failed to confirm choice: %w", err)
		}

		if confirmed {
			return true, domain.SaveTinaCloudConfig(cwd, tinaDir, org.ID, project.ID)
		}
		return false, nil
	}

	// Best-effort check: a failure should not block the manual select path.
	localRepoSlug, _ := repoSlugFromGitConfig(cwd)

	var menuItems []string
	for _, option := range projectOptions {
		menuItems = append(
			menuItems,
			fmt.Sprintf("%s - %s", option.org.Name, option.project.Name),
		)

		if localRepoSlug != "" && repoSlugFromURL(option.project.RepoURL) == localRepoSlug {
			confirmed, err := confirmPrompt(fmt.Sprintf("Continue with %s (%s)?", option.project.Name, option.org.Name))
			if err != nil {
				return false, fmt.Errorf("failed to confirm choice: %w", err)
			}

			if confirmed {
				return true, domain.SaveTinaCloudConfig(cwd, tinaDir, option.org.ID, option.project.ID)
			}
			break
		}
	}

	if len(menuItems) == 0 {
		return false, fmt.Errorf("account has no projects")
	}

	prompt := promptui.Select{
		Label: "We could not guess which project you wanted. Please select a project to configure",
		Items: menuItems,
		Size:  10,
	}
	selectedIndex, _, err := prompt.Run()
	if err != nil {
		return false, fmt.Errorf("failed to select item: %w", err)
	}

	project := projectOptions[selectedIndex].project
	org := projectOptions[selectedIndex].org

	confirmed, err := confirmPrompt(fmt.Sprintf("Continue with %s (%s)?", project.Name, org.Name))
	if err != nil {
		return false, fmt.Errorf("failed to confirm choice: %w", err)
	}

	if confirmed {
		return true, domain.SaveTinaCloudConfig(cwd, tinaDir, org.ID, project.ID)
	}

	return false, nil
}

func confirmPrompt(prompt string) (bool, error) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt: fmt.Sprintf("%s [y/N] ", prompt),
	})
	if err != nil {
		return false, fmt.Errorf("failed to initialize prompt: %w", err)
	}
	defer rl.Close()

	answer, err := rl.Readline()
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

func repoSlugFromURL(repoURL string) string {
	repoURL = strings.TrimSpace(strings.TrimSuffix(repoURL, ".git"))

	if !strings.Contains(repoURL, "://") {
		if _, slug, ok := strings.Cut(repoURL, ":"); ok {
			return slug
		}
	}

	parts := strings.Split(strings.Trim(repoURL, "/"), "/")
	if len(parts) < 2 {
		return ""
	}

	return strings.Join(parts[len(parts)-2:], "/")
}

func repoSlugFromGitConfig(cwd string) (string, error) {
	var localRepoSlug string

	file, err := os.OpenInRoot(path.Join(cwd, ".git"), "config")
	if err != nil {
		return "", fmt.Errorf("failed to open git config: %w", err)
	}
	defer file.Close()

	body, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read tinacloud tokens: %w", err)
	}

	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || strings.TrimSpace(key) != "url" {
			continue
		}

		localRepoSlug = repoSlugFromURL(strings.TrimSpace(value))
	}

	return localRepoSlug, nil
}
