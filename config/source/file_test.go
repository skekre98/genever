package source

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFileSource_Name(t *testing.T) {
	source := &FileSource{}
	expected := "file"
	if got := source.Name(); got != expected {
		t.Errorf("Name() = %v, want %v", got, expected)
	}
}

func TestFileSource_Load(t *testing.T) {
	tests := []struct {
		name        string
		profile     string
		baseContent string
		profContent string
		expected    map[string]any
		wantErr     bool
	}{
		{
			name:    "base file only",
			profile: "dev",
			baseContent: `
app:
  name: test-app
  port: 8080
database:
  host: localhost
  port: 5432
`,
			profContent: "",
			expected: map[string]any{
				"app": map[string]any{
					"name": "test-app",
					"port": 8080,
				},
				"database": map[string]any{
					"host": "localhost",
					"port": 5432,
				},
			},
			wantErr: false,
		},
		{
			name:    "base and profile files - profile replaces top-level keys",
			profile: "prod",
			baseContent: `
app:
  name: test-app
  port: 8080
database:
  host: localhost
  port: 5432
`,
			profContent: `
app:
  port: 9090
database:
  host: prod-db.example.com
  ssl: true
`,
			expected: map[string]any{
				"app": map[string]any{
					"port": 9090,
				},
				"database": map[string]any{
					"host": "prod-db.example.com",
					"ssl":  true,
				},
			},
			wantErr: false,
		},
		{
			name:    "profile adds new top-level keys",
			profile: "test",
			baseContent: `
app:
  name: test-app
`,
			profContent: `
cache:
  enabled: true
`,
			expected: map[string]any{
				"app": map[string]any{
					"name": "test-app",
				},
				"cache": map[string]any{
					"enabled": true,
				},
			},
			wantErr: false,
		},
		{
			name:        "empty base file",
			profile:     "dev",
			baseContent: ``,
			profContent: `
app:
  name: test-app
`,
			expected: map[string]any{
				"app": map[string]any{
					"name": "test-app",
				},
			},
			wantErr: false,
		},
		{
			name:    "profile file missing",
			profile: "nonexistent",
			baseContent: `
app:
  name: test-app
`,
			profContent: "", // Will not create profile file
			expected: map[string]any{
				"app": map[string]any{
					"name": "test-app",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Write base file
			baseFile := filepath.Join(tmpDir, "application.yaml")
			if err := os.WriteFile(baseFile, []byte(tt.baseContent), 0644); err != nil {
				t.Fatalf("Failed to write base file: %v", err)
			}

			// Write profile file if content provided
			if tt.profContent != "" {
				profFile := filepath.Join(tmpDir, "application."+tt.profile+".yaml")
				if err := os.WriteFile(profFile, []byte(tt.profContent), 0644); err != nil {
					t.Fatalf("Failed to write profile file: %v", err)
				}
			}

			// Create FileSource and load
			source := &FileSource{
				BasePath: tmpDir,
				Profile:  tt.profile,
			}
			result, err := source.Load(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Load() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFileSource_Load_BaseFileMissing(t *testing.T) {
	// Create temporary directory but don't create base file
	tmpDir := t.TempDir()

	source := &FileSource{
		BasePath: tmpDir,
		Profile:  "dev",
	}

	_, err := source.Load(context.Background())
	if err == nil {
		t.Error("Load() expected error when base file missing, got nil")
	}
}

func TestFileSource_Load_InvalidYAML(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Write invalid YAML to base file
	baseFile := filepath.Join(tmpDir, "application.yaml")
	invalidYAML := `
app:
  name: test
  port: [invalid
`
	if err := os.WriteFile(baseFile, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write base file: %v", err)
	}

	source := &FileSource{
		BasePath: tmpDir,
		Profile:  "dev",
	}

	_, err := source.Load(context.Background())
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}
}

func TestFileSource_Load_ComplexNesting(t *testing.T) {
	tmpDir := t.TempDir()

	baseContent := `
app:
  server:
    http:
      port: 8080
      timeout: 30s
    grpc:
      port: 9090
  cache:
    redis:
      host: localhost
      port: 6379
database:
  primary:
    host: db1.example.com
    port: 5432
  replica:
    host: db2.example.com
    port: 5432
`

	baseFile := filepath.Join(tmpDir, "application.yaml")
	if err := os.WriteFile(baseFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to write base file: %v", err)
	}

	source := &FileSource{
		BasePath: tmpDir,
		Profile:  "dev",
	}

	result, err := source.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify deeply nested values
	tests := []struct {
		name     string
		path     []string
		expected any
	}{
		{
			name:     "http port",
			path:     []string{"app", "server", "http", "port"},
			expected: 8080,
		},
		{
			name:     "http timeout",
			path:     []string{"app", "server", "http", "timeout"},
			expected: "30s",
		},
		{
			name:     "grpc port",
			path:     []string{"app", "server", "grpc", "port"},
			expected: 9090,
		},
		{
			name:     "redis host",
			path:     []string{"app", "cache", "redis", "host"},
			expected: "localhost",
		},
		{
			name:     "primary db host",
			path:     []string{"database", "primary", "host"},
			expected: "db1.example.com",
		},
		{
			name:     "replica db port",
			path:     []string{"database", "replica", "port"},
			expected: 5432,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := getNestedValueAny(result, tt.path)
			if !reflect.DeepEqual(val, tt.expected) {
				t.Errorf("Load() path %v = %v, want %v", tt.path, val, tt.expected)
			}
		})
	}
}

func TestFileSource_Load_ProfileOverrides(t *testing.T) {
	tmpDir := t.TempDir()

	baseContent := `
app:
  name: base-app
  version: 1.0.0
  debug: false
database:
  host: localhost
  port: 5432
  pool:
    min: 5
    max: 10
`

	profContent := `
app:
  name: prod-app
  debug: true
database:
  host: prod-db.example.com
  pool:
    max: 50
cache:
  enabled: true
`

	baseFile := filepath.Join(tmpDir, "application.yaml")
	if err := os.WriteFile(baseFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to write base file: %v", err)
	}

	profFile := filepath.Join(tmpDir, "application.prod.yaml")
	if err := os.WriteFile(profFile, []byte(profContent), 0644); err != nil {
		t.Fatalf("Failed to write profile file: %v", err)
	}

	source := &FileSource{
		BasePath: tmpDir,
		Profile:  "prod",
	}

	result, err := source.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Note: yaml.Unmarshal replaces entire top-level keys, not deep merging
	// So profile file completely replaces "app" and "database" keys from base
	expected := map[string]any{
		"app": map[string]any{
			"name":  "prod-app",
			"debug": true,
		},
		"database": map[string]any{
			"host": "prod-db.example.com",
			"pool": map[string]any{
				"max": 50,
			},
		},
		"cache": map[string]any{
			"enabled": true,
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Load() = %v, want %v", result, expected)
	}
}

func TestFileSource_Watch(t *testing.T) {
	source := &FileSource{
		BasePath: "/tmp",
		Profile:  "dev",
	}

	// Watch should return nil (no-op for file source)
	if err := source.Watch(context.Background(), nil); err != nil {
		t.Errorf("Watch() returned error: %v", err)
	}
}

// Helper function to get nested value from map (supports any type)
func getNestedValueAny(m map[string]any, path []string) any {
	current := m
	for i, key := range path {
		if i == len(path)-1 {
			return current[key]
		}
		if next, ok := current[key].(map[string]any); ok {
			current = next
		} else {
			return nil
		}
	}
	return nil
}
