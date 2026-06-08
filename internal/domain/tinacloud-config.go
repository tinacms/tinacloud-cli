package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

const DefaultTinaDir = "tina"
const TinaCloudConfigFileName = "tinacloud.json"

type TinaCloudConfig struct {
	OrgID     string `json:"org_id"`
	ProjectID string `json:"project_id"`
}

func SaveTinaCloudConfig(cwd, tinaDir, orgID, projectID string) error {
	_, err := os.Stat(path.Join(cwd, tinaDir))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("failed to find 'tina' directory. please ensure you're running this command inside of a tinacms project.")
		}
		return fmt.Errorf("failed to check tina directory existence: %w", err)
	}

	file, err := os.OpenFile(path.Join(cwd, tinaDir, TinaCloudConfigFileName), os.O_RDWR|os.O_CREATE, 0600) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to create/open tinacloud config: %w", err)
	}
	defer file.Close()

	config := TinaCloudConfig{
		OrgID:     orgID,
		ProjectID: projectID,
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
