package source

import (
	"context"
	"os"
	"reflect"
	"testing"
)

func TestEnvSource_Name(t *testing.T) {
	source := &EnvSource{}
	expected := "env"
	if got := source.Name(); got != expected {
		t.Errorf("Name() = %v, want %v", got, expected)
	}
}

func TestEnvSource_Load(t *testing.T) {
	// Save original environment
	originalEnv := os.Environ()
	defer restoreEnv(originalEnv)

	// Clear environment and set test values
	os.Clearenv()
	testEnvVars := map[string]string{
		"GENEVER_DATABASE_HOST":    "localhost",
		"GENEVER_DATABASE_PORT":    "5432",
		"GENEVER_APP_NAME":         "test-app",
		"GENEVER_APP_SERVER_PORT":  "8080",
		"GENEVER_CACHE_REDIS_HOST": "redis.local",
		"GENEVER_CACHE_REDIS_PORT": "6379",
		"OTHER_VAR":                "should-be-ignored",
		"GENEVER_SIMPLE":           "value",
	}

	for k, v := range testEnvVars {
		os.Setenv(k, v)
	}

	source := &EnvSource{}
	result, err := source.Load(context.Background())

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Verify nested structure
	tests := []struct {
		name     string
		path     []string
		expected string
	}{
		{
			name:     "database host",
			path:     []string{"database", "host"},
			expected: "localhost",
		},
		{
			name:     "database port",
			path:     []string{"database", "port"},
			expected: "5432",
		},
		{
			name:     "app name",
			path:     []string{"app", "name"},
			expected: "test-app",
		},
		{
			name:     "app server port",
			path:     []string{"app", "server", "port"},
			expected: "8080",
		},
		{
			name:     "cache redis host",
			path:     []string{"cache", "redis", "host"},
			expected: "redis.local",
		},
		{
			name:     "simple value",
			path:     []string{"simple"},
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := getNestedValue(result, tt.path)
			if val != tt.expected {
				t.Errorf("Load() path %v = %v, want %v", tt.path, val, tt.expected)
			}
		})
	}

	// Verify non-GENEVER vars are not included
	if val := getNestedValue(result, []string{"other"}); val != "" {
		t.Errorf("Load() should not include non-GENEVER vars, found %v", val)
	}
}

func TestParseEnvLine(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKey   string
		wantValue string
		wantFound bool
	}{
		{
			name:      "valid env var",
			input:     "KEY=value",
			wantKey:   "KEY",
			wantValue: "value",
			wantFound: true,
		},
		{
			name:      "env var with equals in value",
			input:     "KEY=value=with=equals",
			wantKey:   "KEY",
			wantValue: "value=with=equals",
			wantFound: true,
		},
		{
			name:      "empty value",
			input:     "KEY=",
			wantKey:   "KEY",
			wantValue: "",
			wantFound: true,
		},
		{
			name:      "no equals sign",
			input:     "INVALID",
			wantKey:   "",
			wantValue: "",
			wantFound: false,
		},
		{
			name:      "empty string",
			input:     "",
			wantKey:   "",
			wantValue: "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotValue, gotFound := parseEnvLine(tt.input)
			if gotKey != tt.wantKey {
				t.Errorf("parseEnvLine() key = %v, want %v", gotKey, tt.wantKey)
			}
			if gotValue != tt.wantValue {
				t.Errorf("parseEnvLine() value = %v, want %v", gotValue, tt.wantValue)
			}
			if gotFound != tt.wantFound {
				t.Errorf("parseEnvLine() found = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

func TestSetNestedValue(t *testing.T) {
	tests := []struct {
		name     string
		segments []string
		value    string
		want     map[string]any
	}{
		{
			name:     "single segment",
			segments: []string{"key"},
			value:    "value",
			want: map[string]any{
				"key": "value",
			},
		},
		{
			name:     "two segments",
			segments: []string{"database", "host"},
			value:    "localhost",
			want: map[string]any{
				"database": map[string]any{
					"host": "localhost",
				},
			},
		},
		{
			name:     "three segments",
			segments: []string{"app", "server", "port"},
			value:    "8080",
			want: map[string]any{
				"app": map[string]any{
					"server": map[string]any{
						"port": "8080",
					},
				},
			},
		},
		{
			name:     "empty segment handling",
			segments: []string{"app", "", "port"},
			value:    "8080",
			want: map[string]any{
				"app": map[string]any{
					"port": "8080",
				},
			},
		},
		{
			name:     "all empty segments",
			segments: []string{"", ""},
			value:    "value",
			want:     map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make(map[string]any)
			setNestedValue(result, tt.segments, tt.value)
			if !reflect.DeepEqual(result, tt.want) {
				t.Errorf("setNestedValue() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestSetNestedValue_Conflicts(t *testing.T) {
	t.Run("conflict: leaf value exists", func(t *testing.T) {
		m := make(map[string]any)
		// Set a leaf value
		setNestedValue(m, []string{"db"}, "mydb")
		// Try to create nested structure under existing leaf
		setNestedValue(m, []string{"db", "host"}, "localhost")

		// Should preserve original leaf value
		if m["db"] != "mydb" {
			t.Errorf("Expected leaf value to be preserved, got %v", m["db"])
		}
	})

	t.Run("no conflict: nested map exists", func(t *testing.T) {
		m := make(map[string]any)
		// Set first nested value
		setNestedValue(m, []string{"db", "host"}, "localhost")
		// Add another value to same nested map
		setNestedValue(m, []string{"db", "port"}, "5432")

		expected := map[string]any{
			"db": map[string]any{
				"host": "localhost",
				"port": "5432",
			},
		}

		if !reflect.DeepEqual(m, expected) {
			t.Errorf("setNestedValue() = %v, want %v", m, expected)
		}
	})
}

func TestLoadEnvVars_PrefixFiltering(t *testing.T) {
	// Save original environment
	originalEnv := os.Environ()
	defer restoreEnv(originalEnv)

	os.Clearenv()
	os.Setenv("GENEVER_INCLUDED", "yes")
	os.Setenv("NOT_GENEVER", "no")
	os.Setenv("GENEVER_ALSO_INCLUDED", "yes")
	os.Setenv("genever_lowercase", "no") // Should not match - case sensitive

	result := loadEnvVars()

	// Should have 2 entries (both GENEVER_ prefixed)
	if len(result) != 2 {
		t.Errorf("loadEnvVars() returned %d entries, want 2", len(result))
	}

	if result["included"] != "yes" {
		t.Errorf("Expected 'included' key with value 'yes', got %v", result["included"])
	}

	if result["also"] == nil {
		t.Errorf("Expected 'also' key to exist")
	}

	// Verify non-GENEVER vars are excluded
	if _, exists := result["not"]; exists {
		t.Error("Non-GENEVER variable should not be included")
	}
}

func TestLoadEnvVars_CaseConversion(t *testing.T) {
	// Save original environment
	originalEnv := os.Environ()
	defer restoreEnv(originalEnv)

	os.Clearenv()
	os.Setenv("GENEVER_UPPERCASE_KEY", "value")
	os.Setenv("GENEVER_MixedCase_Key", "value2")

	result := loadEnvVars()

	// Keys should be lowercase
	if _, exists := result["uppercase"]; !exists {
		t.Error("Expected lowercase 'uppercase' key")
	}

	if _, exists := result["mixedcase"]; !exists {
		t.Error("Expected lowercase 'mixedcase' key")
	}
}

func TestLoadEnvVars_ComplexNesting(t *testing.T) {
	// Save original environment
	originalEnv := os.Environ()
	defer restoreEnv(originalEnv)

	os.Clearenv()
	os.Setenv("GENEVER_A_B_C_D", "deep")
	os.Setenv("GENEVER_A_B_X", "branch1")
	os.Setenv("GENEVER_A_Y", "branch2")

	result := loadEnvVars()

	// Verify deep nesting
	if val := getNestedValue(result, []string{"a", "b", "c", "d"}); val != "deep" {
		t.Errorf("Deep nested value = %v, want 'deep'", val)
	}

	// Verify sibling values
	if val := getNestedValue(result, []string{"a", "b", "x"}); val != "branch1" {
		t.Errorf("Sibling value = %v, want 'branch1'", val)
	}

	if val := getNestedValue(result, []string{"a", "y"}); val != "branch2" {
		t.Errorf("Parent sibling value = %v, want 'branch2'", val)
	}
}

// Helper function to get nested value from map
func getNestedValue(m map[string]any, path []string) string {
	current := m
	for i, key := range path {
		if i == len(path)-1 {
			if val, ok := current[key].(string); ok {
				return val
			}
			return ""
		}
		if next, ok := current[key].(map[string]any); ok {
			current = next
		} else {
			return ""
		}
	}
	return ""
}

// Helper function to restore environment
func restoreEnv(originalEnv []string) {
	os.Clearenv()
	for _, env := range originalEnv {
		if key, val, found := parseEnvLine(env); found {
			os.Setenv(key, val)
		}
	}
}
