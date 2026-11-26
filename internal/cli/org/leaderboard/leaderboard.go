package leaderboard

import (
	"cmp"
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"text/tabwriter"
	"time"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"

	"tina.io/tinacloud-cli/internal/httpclient"
	"tina.io/tinacloud-cli/internal/requests"
	"tina.io/tinacloud-cli/internal/utils"
)

type leaderboardJSONOutput struct {
	LoggedIn bool                      `json:"loggedIn"`
	Entries  []domain.LeaderboardEntry `json:"entries"`
}

func Command(cfg config.Config) *command.Command {
	flagSet := flag.NewFlagSet("leaderboard", flag.ContinueOnError)
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var orgIDFlag = flagSet.String("org-id", "", "TinaCloud organization ID")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")
	var allFlag = flagSet.Bool("all", false, "Show all users")

	now := time.Now()
	var toFlag = flagSet.String("to", now.Format(time.DateOnly), "YYYY-MM-DD")
	var fromFlag = flagSet.String("from", now.AddDate(0, 0, 1-now.Day()).Format(time.DateOnly), "YYYY-MM-DD")

	return &command.Command{
		Name:        "leaderboard",
		Description: "TODO",
		Flags:       flagSet,
		Run: func(args []string) error {
			out, err := runLeaderboard(cfg, auth.TokensFromCache, *orgIDFlag, *tinaDirFlag, *toFlag, *fromFlag)
			if *jsonFlag {
				if !*allFlag {
					out.Entries = slices.DeleteFunc(out.Entries, func(a domain.LeaderboardEntry) bool {
						return a.Count <= 0
					})
				}
				return utils.WriteJSONResult(out, err)
			}
			if err != nil {
				return err
			}

			writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			fmt.Fprintf(writer, "Leaderboard %s to %s\n", *fromFlag, *toFlag)
			for _, entry := range out.Entries {
				if !*allFlag && entry.Count == 0 {
					continue
				}

				name := entry.Name
				if len(name) == 0 {
					if entry.FormerMember {
						name = "Former member"
					} else {
						name = "Unknown member"
					}
				}
				fmt.Fprintf(writer, "\t- %d.\t%s\t%d\n", entry.Rank, name, entry.Count)
			}

			err = writer.Flush()
			if err != nil {
				return fmt.Errorf("failed to flush writer: %w", err)
			}

			return nil
		},
	}
}

func runLeaderboard(cfg config.Config, loadTokens auth.TokenLoader, orgID, tinaDir, to, from string) (leaderboardJSONOutput, error) {
	if orgID == "" {
		if cwd, err := os.Getwd(); err == nil {
			if tinacloudConfig, err := domain.TinaCloudConfigFromDisk(cwd, tinaDir); err == nil {
				orgID = tinacloudConfig.OrgID
			}
		}

		if orgID == "" {
			return leaderboardJSONOutput{}, errors.New("organization ID is required")
		}
	}

	tokens, err := loadTokens(cfg)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) || errors.Is(err, httpclient.ErrUnauthorized) {
			return leaderboardJSONOutput{}, err
		}
		return leaderboardJSONOutput{}, fmt.Errorf("failed to initialize auth client: %w", err)
	}

	entries, err := requests.GetLeaderboard(cfg.TinaCloud.ContentEndpoint, tokens.AccessToken, orgID, to, from)
	if err != nil {
		return leaderboardJSONOutput{}, err
	}

	slices.SortFunc(entries, func(a, b domain.LeaderboardEntry) int {
		return cmp.Compare(a.Rank, b.Rank)
	})

	return leaderboardJSONOutput{
		LoggedIn: true,
		Entries:  entries,
	}, nil
}
