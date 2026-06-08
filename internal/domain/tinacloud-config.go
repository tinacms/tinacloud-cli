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
