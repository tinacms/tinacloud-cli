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
	LoggedIn      bool                          `json:"loggedIn"`
	Organizations []domain.OrganizationWithRole `json:"organizations"`
}

func List(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("orglist", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli org list [flags]")
		fmt.Fprintln(os.Stderr, "\nList organizations")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	_ = flagSet.Parse(args)

	out, err := runList(cfg, auth.TokensFromCache)
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

	fmt.Println("Organizations:")
	for _, org := range out.Organizations {
		fmt.Printf("\t- %s\n", org.String())
	}

	return nil
}

func runList(cfg config.Config, loadTokens auth.TokenLoader) (listJSONOutput, error) {
	tokens, err := loadTokens(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, auth.ErrUnauthorized) {
			return listJSONOutput{LoggedIn: false}, nil
		}
		return listJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/organizations", cfg.TinaCloud.Endpoint), nil)
	if err != nil {
		return listJSONOutput{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return listJSONOutput{}, fmt.Errorf("failed to get organizations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return listJSONOutput{}, auth.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return listJSONOutput{}, fmt.Errorf("failed to read get organizations response: %w", err)
	}

	if resp.StatusCode > 299 {
		return listJSONOutput{}, fmt.Errorf("failed to get organizations. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var organizations []domain.OrganizationWithRole
	if err := json.Unmarshal(body, &organizations); err != nil {
		return listJSONOutput{}, fmt.Errorf("failed to unmarshal organizations from get request: %w", err)
	}

	return listJSONOutput{
		LoggedIn:      true,
		Organizations: organizations,
	}, nil
}
