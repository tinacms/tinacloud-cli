package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
)

const DefaultTinaDir = "tina"
const TinaCloudConfigFileName = "tinacloud.json"

// ErrNotTinaProject indicates the tina directory was not found, meaning the
// command is not being run inside a tinacms project.
var ErrNotTinaProject = errors.New("not a tinacms project: 'tina' directory not found")

type TinaCloudConfig struct {
	OrgID     string `json:"org_id"`
	ProjectID string `json:"project_id"`
}

func SaveTinaCloudConfig(cwd, tinaDir, orgID, projectID string) error {
	_, err := os.Stat(path.Join(cwd, tinaDir))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotTinaProject
		}
		return fmt.Errorf("failed to check tina directory existence: %w", err)
	}

	config := TinaCloudConfig{
		OrgID:     orgID,
		ProjectID: projectID,
	}

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal tinacloud config: %w", err)
	}

	err = os.WriteFile(path.Join(cwd, tinaDir, TinaCloudConfigFileName), data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write tinacloud config: %w", err)
	}

	return nil
}

// It is up to the caller to validate whether the tokens within the struct are
// valid
func TinaCloudConfigFromDisk(cwd, tinaDir string) (TinaCloudConfig, error) {
	file, err := os.OpenInRoot(path.Join(cwd, tinaDir), TinaCloudConfigFileName)
	if err != nil {
		return TinaCloudConfig{}, fmt.Errorf("failed to open tinacloud config: %w", err)
	}

	const bufSize int = 10240 // 10KB
	buf := make([]byte, bufSize)
	bytes, err := file.Read(buf)
	if err != nil {
		return TinaCloudConfig{}, fmt.Errorf("failed to read tinacloud tokens: %w", err)
	}

	var config TinaCloudConfig
	err = json.Unmarshal(buf[:bytes], &config)
	if err != nil {
		return TinaCloudConfig{}, fmt.Errorf("failed to parse tinacloud config")
	}

	return config, nil
}
