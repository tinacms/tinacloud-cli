package configure

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"

	"tina.io/tinacloud-cli/internal/domain"
)

func Configure(args []string) error {
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

	_, err = os.Stat(path.Join(cwd, *tinaDirFlag))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("failed to find 'tina' directory. please ensure you're running this command inside of a tinacms project.")
		}
		return fmt.Errorf("failed to check tina directory existence: %w", err)
	}

	file, err := os.OpenFile(path.Join(cwd, *tinaDirFlag, domain.TinaCloudConfigFileName), os.O_RDWR|os.O_CREATE, 0600) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to create/open tinacloud config: %w", err)
	}
	defer file.Close()

	config := domain.TinaCloudConfig{
		OrgID:     *orgIDFlag,
		ProjectID: *projectIDFlag,
	}

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal tinacloud config: %w", err)
	}

	bytes, err := file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write config to tinacloud config: %w", err)
	}
	err = file.Truncate(int64(bytes))
	if err != nil {
		return fmt.Errorf("failed to truncate after write: %w", err)
	}

	fmt.Printf("Wrote tinacloud config to %s\n", path.Join(cwd, *tinaDirFlag, domain.TinaCloudConfigFileName))

	return nil
}
