package leaderboard

import (
	"cmp"
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
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
	Landing  *domain.Landing           `json:"landing,omitempty"`
}

func Command(cfg config.Config, target requests.LeaderboardTarget) *command.Command {
	idFlagName, targetName := "org-id", "organization"
	if target == requests.Project {
		idFlagName, targetName = "project-id", "project"
	}
	flagSet := flag.NewFlagSet("leaderboard", flag.ContinueOnError)
	var jsonFlag = flagSet.Bool("json", false, "Print the result as JSON")
	var idFlag = flagSet.String(idFlagName, "", "TinaCloud "+targetName+" ID")
	var tinaDirFlag = flagSet.String("tina-dir", domain.DefaultTinaDir, "Path to tina folder")
	var allFlag = flagSet.Bool("all", false, "Include users with zero activity")
	var periodFlag = flagSet.String("period", "", "UTC calendar period: this-month, last-month, 3-months, or 12-months (includes current month)")

	now := time.Now().UTC()
	var toFlag = flagSet.String("to", now.Format(time.DateOnly), "End date (YYYY-MM-DD; default today in UTC)")
	var fromFlag = flagSet.String("from", now.AddDate(0, 0, 1-now.Day()).Format(time.DateOnly), "Start date (YYYY-MM-DD; default first day of this month in UTC)")

	return &command.Command{
		Name:        "leaderboard",
		Description: "Rank " + targetName + " users by activity count",
		Flags:       flagSet,
		Run: func(args []string) error {
			var fromSet, toSet bool
			flagSet.Visit(func(f *flag.Flag) {
				fromSet = fromSet || f.Name == "from"
				toSet = toSet || f.Name == "to"
			})
			if *periodFlag != "" {
				if fromSet || toSet {
					return errors.New("--period cannot be combined with --from or --to")
				}
				from, to, err := periodDates(*periodFlag, now)
				if err != nil {
					return err
				}
				*fromFlag, *toFlag = from, to
			}
			if fromSet != toSet {
				return errors.New("--from and --to must be used together")
			}
			from, err := time.Parse(time.DateOnly, *fromFlag)
			if err != nil {
				return fmt.Errorf("invalid --from date %q: expected YYYY-MM-DD", *fromFlag)
			}
			to, err := time.Parse(time.DateOnly, *toFlag)
			if err != nil {
				return fmt.Errorf("invalid --to date %q: expected YYYY-MM-DD", *toFlag)
			}
			if from.After(to) {
				return errors.New("--from must not be after --to")
			}
			out, err := runLeaderboard(cfg, auth.TokensFromCache, *idFlag, *tinaDirFlag, *toFlag, *fromFlag, target)
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
			if out.Landing != nil {
				var names []string
				for _, contributor := range out.Landing.Contributors {
					name := "No name set"
					if contributor.Name != nil {
						name = *contributor.Name
					} else if contributor.FormerMember {
						name = "Former member"
					}
					names = append(names, name)
				}
				fmt.Fprintf(writer, "Last landed on default branch at %s by %s\n", time.UnixMilli(out.Landing.At).UTC().Format(time.RFC3339), strings.Join(names, ", "))
			}
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

func periodDates(period string, now time.Time) (string, string, error) {
	parsedPeriod, err := domain.LeaderboardPeriodFromString(period)
	if err != nil {
		return "", "", err
	}
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now
	switch parsedPeriod {
	case domain.LastMonth:
		end = start.AddDate(0, 0, -1)
		start = start.AddDate(0, -1, 0)
	case domain.ThreeMonths:
		start = start.AddDate(0, -2, 0)
	case domain.TwelveMonths:
		start = start.AddDate(0, -11, 0)
	}
	return start.Format(time.DateOnly), end.Format(time.DateOnly), nil
}

func runLeaderboard(cfg config.Config, loadTokens auth.TokenLoader, id, tinaDir, to, from string, target requests.LeaderboardTarget) (leaderboardJSONOutput, error) {
	if id == "" {
		if cwd, err := os.Getwd(); err == nil {
			if tinacloudConfig, err := domain.TinaCloudConfigFromDisk(cwd, tinaDir); err == nil {
				if target == requests.Project {
					id = tinacloudConfig.ProjectID
				} else {
					id = tinacloudConfig.OrgID
				}
			}
		}

		if id == "" {
			if target == requests.Project {
				return leaderboardJSONOutput{}, errors.New("project ID is required")
			}
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

	leaderboard, err := requests.GetLeaderboard(cfg.TinaCloud.ContentEndpoint, tokens.AccessToken, id, to, from, target)
	if err != nil {
		return leaderboardJSONOutput{}, err
	}

	slices.SortFunc(leaderboard.Entries, func(a, b domain.LeaderboardEntry) int {
		return cmp.Compare(a.Rank, b.Rank)
	})

	return leaderboardJSONOutput{
		LoggedIn: true,
		Entries:  leaderboard.Entries,
		Landing:  leaderboard.Landing,
	}, nil
}
