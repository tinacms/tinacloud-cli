package list

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
)

type listJSONOutput struct {
	LoggedIn bool                           `json:"loggedIn"`
	Projects []domain.GetV2ProjectsResponse `json:"projects"`
}

func List(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("projectlist", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli project list [flags]")
		fmt.Fprintln(os.Stderr, "\nList projects")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	_ = flagSet.Parse(args)

	if *orgIDFlag == "" {
		return fmt.Errorf("organization ID is required")
	}

	out, err := runList(cfg, *orgIDFlag)
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

	fmt.Println("Projects:")
	for _, project := range out.Projects {
		fmt.Printf("\t- %s\n", project.String())
	}

	return nil
}

func runList(cfg config.Config, orgIDFlag string) (listJSONOutput, error) {
	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return listJSONOutput{LoggedIn: false}, nil
		}
		return listJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/organizations/%s/apps", cfg.TinaCloud.Endpoint, orgIDFlag), nil)
	if err != nil {
		return listJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return listJSONOutput{}, fmt.Errorf("failed to get projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return listJSONOutput{}, auth.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return listJSONOutput{}, fmt.Errorf("failed to read get projects response: %w", err)
	}

	if resp.StatusCode > 299 {
		return listJSONOutput{}, fmt.Errorf("failed to get projects. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var projects []domain.GetV2ProjectsResponse
	if err := json.Unmarshal(body, &projects); err != nil {
		return listJSONOutput{}, fmt.Errorf("failed to unmarshal projects from get request: %w", err)
	}

	return listJSONOutput{
		LoggedIn: true,
		Projects: projects,
	}, nil
}
