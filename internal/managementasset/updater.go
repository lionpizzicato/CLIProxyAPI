package managementasset

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/lionpizzicato/CLIProxyAPI/v6/internal/config"
	"github.com/lionpizzicato/CLIProxyAPI/v6/internal/util"
	log "github.com/sirupsen/logrus"
)

const managementAssetName = "management.html"

// ManagementFileName exposes the control panel asset filename.
const ManagementFileName = managementAssetName

//go:embed management.html
var embeddedManagementHTML []byte

// SetCurrentConfig preserves the historical package API. The management UI is now bundled
// with the binary, so no runtime config tracking is required here.
func SetCurrentConfig(_ *config.Config) {}

// StartAutoUpdater preserves the historical package API. The management UI is now bundled
// with the binary, so no runtime download or sync loop is required.
func StartAutoUpdater(_ context.Context, _ string) {
	log.Debug("management asset auto-updater skipped: using bundled asset")
}

// StaticDir resolves the directory that stores the management control panel asset.
// This remains available for explicit MANAGEMENT_STATIC_PATH overrides.
func StaticDir(configFilePath string) string {
	if override := strings.TrimSpace(os.Getenv("MANAGEMENT_STATIC_PATH")); override != "" {
		cleaned := filepath.Clean(override)
		if strings.EqualFold(filepath.Base(cleaned), managementAssetName) {
			return filepath.Dir(cleaned)
		}
		return cleaned
	}

	if writable := util.WritablePath(); writable != "" {
		return filepath.Join(writable, "static")
	}

	configFilePath = strings.TrimSpace(configFilePath)
	if configFilePath == "" {
		return ""
	}

	base := filepath.Dir(configFilePath)
	fileInfo, err := os.Stat(configFilePath)
	if err == nil && fileInfo.IsDir() {
		base = configFilePath
	}

	return filepath.Join(base, "static")
}

// FilePath resolves the absolute path to the management control panel asset.
func FilePath(configFilePath string) string {
	if override := strings.TrimSpace(os.Getenv("MANAGEMENT_STATIC_PATH")); override != "" {
		cleaned := filepath.Clean(override)
		if strings.EqualFold(filepath.Base(cleaned), managementAssetName) {
			return cleaned
		}
		return filepath.Join(cleaned, ManagementFileName)
	}

	dir := StaticDir(configFilePath)
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, ManagementFileName)
}

// LoadManagementHTML returns the bundled management page. When MANAGEMENT_STATIC_PATH is set,
// an on-disk override is preferred.
func LoadManagementHTML(configFilePath string) ([]byte, error) {
	if override := strings.TrimSpace(os.Getenv("MANAGEMENT_STATIC_PATH")); override != "" {
		data, err := os.ReadFile(FilePath(configFilePath))
		if err == nil {
			return data, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	if len(embeddedManagementHTML) == 0 {
		return nil, os.ErrNotExist
	}

	return embeddedManagementHTML, nil
}

// EnsureLatestManagementHTML preserves the historical package API. Instead of downloading
// assets, it materializes the bundled HTML only when an explicit static path is requested.
func EnsureLatestManagementHTML(_ context.Context, staticDir string, _ string, _ string) bool {
	if len(embeddedManagementHTML) == 0 {
		return false
	}

	targetPath := ""
	if override := strings.TrimSpace(os.Getenv("MANAGEMENT_STATIC_PATH")); override != "" {
		targetPath = FilePath("")
	} else if trimmed := strings.TrimSpace(staticDir); trimmed != "" {
		targetPath = filepath.Join(trimmed, managementAssetName)
	}

	if targetPath == "" {
		return true
	}

	if err := writeEmbeddedAsset(targetPath); err != nil {
		log.WithError(err).Warn("failed to materialize bundled management asset")
		return false
	}

	return true
}

func writeEmbeddedAsset(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	if existing, err := os.ReadFile(path); err == nil {
		if bytes.Equal(existing, embeddedManagementHTML) {
			return nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return atomicWriteFile(path, embeddedManagementHTML)
}

func atomicWriteFile(path string, data []byte) error {
	tmpFile, err := os.CreateTemp(filepath.Dir(path), "management-*.html")
	if err != nil {
		return err
	}

	tmpName := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
	}()

	if _, err = tmpFile.Write(data); err != nil {
		return err
	}

	if err = tmpFile.Chmod(0o644); err != nil {
		return err
	}

	if err = tmpFile.Close(); err != nil {
		return err
	}

	if err = os.Rename(tmpName, path); err != nil {
		return err
	}

	return nil
}
