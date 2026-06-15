package domain

import (
	"errors"
	"os"
	"path"
	"testing"
)

func TestSaveAndLoadTinaCloudConfig(t *testing.T) {
	cwd := t.TempDir()
	if err := os.Mkdir(path.Join(cwd, DefaultTinaDir), 0700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	if err := SaveTinaCloudConfig(cwd, DefaultTinaDir, "org-1", "project-1"); err != nil {
		t.Fatalf("SaveTinaCloudConfig() = %v, want nil error", err)
	}

	got, err := TinaCloudConfigFromDisk(cwd, DefaultTinaDir)
	if err != nil {
		t.Fatalf("TinaCloudConfigFromDisk() = %v, want nil error", err)
	}
	want := TinaCloudConfig{OrgID: "org-1", ProjectID: "project-1"}
	if got != want {
		t.Errorf("TinaCloudConfigFromDisk() = %+v, want %+v", got, want)
	}
}

func TestSaveTinaCloudConfigMissingTinaDir(t *testing.T) {
	cwd := t.TempDir()

	err := SaveTinaCloudConfig(cwd, DefaultTinaDir, "org-1", "project-1")
	if !errors.Is(err, ErrNotTinaProject) {
		t.Fatalf("SaveTinaCloudConfig() = %v, want ErrNotTinaProject", err)
	}
}
