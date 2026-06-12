package configure

import (
	"flag"
	"fmt"
	"os"
	"path"

	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/posthog"
)

func Configure(args []string, track posthog.Tracker) error {
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

	if *orgIDFlag == "" || *projectIDFlag == "" {
		return fmt.Errorf("both --org-id and --project-id must be passed")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	err = domain.SaveTinaCloudConfig(cwd, *tinaDirFlag, *orgIDFlag, *projectIDFlag)
	if err != nil {
		return err
	}

	// Both IDs are passed explicitly today; auto-detect and
	// only-one-project flows don't exist yet.
	track(posthog.UserConfigured(posthog.ConfigMethodManualSelect))

	fmt.Printf("Wrote tinacloud config to %s\n", path.Join(cwd, *tinaDirFlag, domain.TinaCloudConfigFileName))

	return nil
}
