package status

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

type statusJSONOutput struct {
	LoggedIn bool                        `json:"loggedIn"`
	Project  domain.GetV2ProjectResponse `json:"project"`
}

func Status(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("projectstatus", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli project status [flags]")
		fmt.Fprintln(os.Stderr, "\nList a projects status")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var projectIDFlag = flagSet.String("project-id", "", "TinaCloud project ID")
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	_ = flagSet.Parse(args)

	if *projectIDFlag == "" {
		return fmt.Errorf("project ID is required")
	}

	out, err := runStatus(cfg, *projectIDFlag)
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

	fmt.Printf("%s\n", out.Project.String())

	return nil
}

func runStatus(cfg config.Config, projectIDFlag string) (statusJSONOutput, error) {
	tokens, err := auth.TokensFromCache(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return statusJSONOutput{LoggedIn: false}, nil
		}
		return statusJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/apps/%s", cfg.TinaCloud.Endpoint, projectIDFlag), nil)
	if err != nil {
		return statusJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return statusJSONOutput{}, fmt.Errorf("failed to get projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return statusJSONOutput{}, auth.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return statusJSONOutput{}, fmt.Errorf("failed to read get projects response: %w", err)
	}

	if resp.StatusCode > 299 {
		return statusJSONOutput{}, fmt.Errorf("failed to get projects. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var project domain.GetV2ProjectResponse
	if err := json.Unmarshal(body, &project); err != nil {
		return statusJSONOutput{}, fmt.Errorf("failed to unmarshal project from get request: %w", err)
	}

	return statusJSONOutput{
		LoggedIn: true,
		Project:  project,
	}, nil
}
