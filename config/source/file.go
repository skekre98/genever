package source

import (
	"context"
	"os"
	"path/filepath"

	"github.com/skekre98/genever/config"
	"gopkg.in/yaml.v3"
)

// FileSource loads configuration from YAML files on the filesystem.
//
// FileSource supports a base configuration file and an optional profile-specific
// overlay file. Both .yaml and .yml extensions are supported.
//
// File loading order:
//  1. Load application.yaml (or application.yml) from BasePath
//  2. If Profile is set, load application.{profile}.yaml as an overlay
//
// The profile file's values override the base file's values at the top level.
// Note: YAML unmarshaling replaces entire top-level keys rather than deep merging.
//
// Example directory structure:
//
//	configs/
//	  application.yaml      # Base configuration
//	  application.dev.yaml  # Development profile
//	  application.prod.yaml # Production profile
//
// Usage:
//
//	source := &FileSource{
//	    BasePath: "configs",
//	    Profile:  "prod",
//	}
type FileSource struct {
	// BasePath is the directory containing the configuration files.
	// The base file (application.yaml) must exist in this directory.
	BasePath string

	// Profile specifies an optional configuration profile.
	// If set, application.{Profile}.yaml will be loaded as an overlay.
	// If the profile file doesn't exist, it's silently ignored.
	Profile string
}

// Name returns the identifier for this source.
func (f *FileSource) Name() string { return "file" }

// Load reads YAML configuration files from the filesystem.
//
// Loads the base file (application.yaml or application.yml) and optionally
// overlays a profile-specific file if Profile is set.
//
// The context is currently not used but is included for future support of
// cancellation and timeouts.
//
// Returns os.ErrNotExist if the base file is not found.
// Returns a YAML parsing error if the files are malformed.
func (f *FileSource) Load(ctx context.Context) (map[string]any, error) {
	// Try both .yaml and .yml extensions for the base file
	baseFile := findYAMLFile(f.BasePath, "application")
	if baseFile == "" {
		return nil, os.ErrNotExist
	}

	data := map[string]any{}
	if err := readYAML(baseFile, data); err != nil {
		return nil, err
	}

	// Try to load profile-specific config if profile is set
	if f.Profile != "" {
		profileFile := findYAMLFile(f.BasePath, "application."+f.Profile)
		if profileFile != "" {
			_ = readYAML(profileFile, data)
		}
	}

	return data, nil
}

// findYAMLFile looks for a file with either .yaml or .yml extension
func findYAMLFile(dir, basename string) string {
	for _, ext := range []string{".yaml", ".yml"} {
		path := filepath.Join(dir, basename+ext)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// Watch is not implemented for FileSource.
// Returns nil immediately, indicating that file watching is not supported.
//
// To enable automatic reloading when files change, consider using a file
// watcher library like fsnotify and implementing Watch accordingly.
func (f *FileSource) Watch(ctx context.Context, ch chan<- config.Event) error { return nil }

func readYAML(path string, out map[string]any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, &out)
}
