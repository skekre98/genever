package source

import (
	"context"
	"os"
	"reflect"
	"testing"
)

func TestCLISource_Load(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string]any
	}{
		{
			name: "simple flags",
			args: []string{"--port=8080", "--host=localhost"},
			expected: map[string]any{
				"port": "8080",
				"host": "localhost",
			},
		},
		{
			name: "dot notation flags",
			args: []string{"--http.port=8080", "--database.host=localhost", "--database.port=5432"},
			expected: map[string]any{
				"http": map[string]any{
					"port": "8080",
				},
				"database": map[string]any{
					"host": "localhost",
					"port": "5432",
				},
			},
		},
		{
			name: "space-separated values",
			args: []string{"--http.port", "8080", "--database.host", "localhost"},
			expected: map[string]any{
				"http": map[string]any{
					"port": "8080",
				},
				"database": map[string]any{
					"host": "localhost",
				},
			},
		},
		{
			name: "mixed formats",
			args: []string{"--http.port=8080", "--database.host", "localhost", "--cache.enabled=true"},
			expected: map[string]any{
				"http": map[string]any{
					"port": "8080",
				},
				"database": map[string]any{
					"host": "localhost",
				},
				"cache": map[string]any{
					"enabled": "true",
				},
			},
		},
		{
			name: "deeply nested",
			args: []string{"--app.server.http.port=8080", "--app.server.http.timeout=30s"},
			expected: map[string]any{
				"app": map[string]any{
					"server": map[string]any{
						"http": map[string]any{
							"port":    "8080",
							"timeout": "30s",
						},
					},
				},
			},
		},
		{
			name: "single dash flags",
			args: []string{"-http.port=8080", "-database.host=localhost"},
			expected: map[string]any{
				"http": map[string]any{
					"port": "8080",
				},
				"database": map[string]any{
					"host": "localhost",
				},
			},
		},
		{
			name: "empty values ignored",
			args: []string{"--http.port=", "--database.host=localhost"},
			expected: map[string]any{
				"database": map[string]any{
					"host": "localhost",
				},
			},
		},
		{
			name: "non-flag arguments ignored",
			args: []string{"command", "--http.port=8080", "argument", "--database.host=localhost"},
			expected: map[string]any{
				"http": map[string]any{
					"port": "8080",
				},
				"database": map[string]any{
					"host": "localhost",
				},
			},
		},
		{
			name:     "no flags",
			args:     []string{"command", "arg1", "arg2"},
			expected: map[string]any{},
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original args and restore after test
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Set test arguments
			os.Args = append([]string{"test"}, tt.args...)

			// Create CLI source and load config
			source := &CLISource{}
			result, err := source.Load(context.Background())
			if err != nil {
				t.Fatalf("Load() returned error: %v", err)
			}

			// Compare results
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Load() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCLISource_Name(t *testing.T) {
	source := &CLISource{}
	if name := source.Name(); name != "cli" {
		t.Errorf("Name() = %v, want %v", name, "cli")
	}
}

func TestCLISource_Watch(t *testing.T) {
	source := &CLISource{}
	ch := make(chan any)

	// Watch should return nil (no-op for CLI source)
	if err := source.Watch(context.Background(), nil); err != nil {
		t.Errorf("Watch() returned error: %v", err)
	}

	close(ch)
}

func TestNormalizeArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "double dash unchanged",
			args:     []string{"--flag=value", "--another-flag"},
			expected: []string{"--flag=value", "--another-flag"},
		},
		{
			name:     "single dash long flag converted",
			args:     []string{"-flag=value", "-another.flag"},
			expected: []string{"--flag=value", "--another.flag"},
		},
		{
			name:     "single char flag unchanged",
			args:     []string{"-v", "-h"},
			expected: []string{"-v", "-h"},
		},
		{
			name:     "mixed arguments",
			args:     []string{"command", "--flag1=value", "-flag2=value", "arg", "-v"},
			expected: []string{"command", "--flag1=value", "--flag2=value", "arg", "-v"},
		},
		{
			name:     "single dash alone unchanged",
			args:     []string{"-"},
			expected: []string{"-"},
		},
		{
			name:     "non-flag arguments unchanged",
			args:     []string{"arg1", "arg2", "arg3"},
			expected: []string{"arg1", "arg2", "arg3"},
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeArgs(tt.args)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("normalizeArgs(%v) = %v, want %v", tt.args, result, tt.expected)
			}
		})
	}
}

func TestExtractFlagName(t *testing.T) {
	tests := []struct {
		arg      string
		expected string
	}{
		{"--flag=value", "flag"},
		{"--flag", "flag"},
		{"-flag=value", "flag"},
		{"-flag", "flag"},
		{"--http.port=8080", "http.port"},
		{"--database.host", "database.host"},
		{"---flag", "flag"}, // Multiple dashes
		{"--=value", ""},    // Just equals
		{"--", ""},          // Just dashes
		{"", ""},            // Empty
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			result := extractFlagName(tt.arg)
			if result != tt.expected {
				t.Errorf("extractFlagName(%q) = %q, want %q", tt.arg, result, tt.expected)
			}
		})
	}
}
