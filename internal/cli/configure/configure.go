package configure

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/chzyer/readline"
	"github.com/manifoldco/promptui"
	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/posthog"
	"tina.io/tinacloud-cli/internal/requests"
)

type projectOption struct {
	project domain.GetV2ProjectsResponse
	org     domain.Organization
}

func Command(cfg config.Config, track posthog.Tracker) *command.Command {
	flagSet := flag.NewFlagSet("configure", flag.ContinueOnError)
	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")

	return &command.Command{
		Name:        "configure",
		Description: "Configure a local Tina project",
		Flags:       flagSet,
		Run: func(args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			saved, err := runConfigure(cfg, cwd, *orgIDFlag, *projectIDFlag, *tinaDirFlag, auth.TokensFromCache, track)
			if err != nil {
				if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
					fmt.Println("You are not currently logged in")
					return nil
				}
				return err
			}

			if saved {
				fmt.Printf("Wrote tinacloud config to %s\n", path.Join(cwd, *tinaDirFlag, domain.TinaCloudConfigFileName))
			} else {
				fmt.Println("Did not write to tinacloud config")
			}

			return nil
		},
	}
}

func runConfigure(cfg config.Config, cwd, orgID, projectID, tinaDir string, loadTokens auth.TokenLoader, track posthog.Tracker) (bool, error) {
	if orgID != "" || projectID != "" {
		if orgID == "" || projectID == "" {
			return false, fmt.Errorf("both --org-id and --project-id must be passed")
		}
		return saveAndTrack(cwd, tinaDir, orgID, projectID, posthog.ConfigMethodManualSelect, track)
	}

	tokens, err := loadTokens(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return false, err
		}
		return false, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	organizations, err := requests.GetOrganizations(cfg.TinaCloud.Endpoint, tokens.AccessToken)
	if err != nil {
		return false, err
	}

	var projectOptions []projectOption
	for _, org := range organizations {
		projects, err := requests.GetProjects(cfg.TinaCloud.Endpoint, tokens.AccessToken, org.Organization.ID)
		if err != nil {
			return false, err
		}

		for _, project := range projects {
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
			return saveAndTrack(cwd, tinaDir, org.ID, project.ID, posthog.ConfigMethodOnlyOneProject, track)
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
				return saveAndTrack(cwd, tinaDir, option.org.ID, option.project.ID, posthog.ConfigMethodAutoDetect, track)
			}
			continue
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
		return saveAndTrack(cwd, tinaDir, org.ID, project.ID, posthog.ConfigMethodManualSelect, track)
	}

	return false, nil
}

// saveAndTrack writes the config, then emits the user-configured event only on success.
func saveAndTrack(cwd, tinaDir, orgID, projectID string, method posthog.ConfigMethod, track posthog.Tracker) (bool, error) {
	if err := domain.SaveTinaCloudConfig(cwd, tinaDir, orgID, projectID); err != nil {
		return false, err
	}

	track(posthog.UserConfigured(method))
	return true, nil
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
		return "", fmt.Errorf("failed to read git config: %w", err)
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
