package source

import (
	"context"
	"os"
	"strings"

	"github.com/skekre98/genever/config"
)

// ENV_PREFIX is the required prefix for environment variables.
// Only variables starting with this prefix are loaded.
const ENV_PREFIX = "GENEVER_"

// EnvSource loads configuration from environment variables.
//
// EnvSource filters environment variables by prefix and converts them into
// a nested map structure using underscores as delimiters.
//
// Environment variable format:
//   - Must start with "GENEVER_" prefix
//   - Remaining parts are split by underscores
//   - Converted to lowercase
//   - Creates nested maps for hierarchical structure
//
// Examples:
//
//	GENEVER_DATABASE_HOST=localhost
//	  -> {database: {host: "localhost"}}
//
//	GENEVER_SERVER_PORT=8080
//	  -> {server: {port: "8080"}}
//
//	GENEVER_APP_NAME=myapp
//	  -> {app: {name: "myapp"}}
//
// All values are returned as strings. Type conversion happens during binding.
//
// Conflict handling:
// If a leaf value already exists, nested values cannot be created at that path.
// For example, setting both GENEVER_DB=value and GENEVER_DB_HOST=localhost
// will preserve the first one and skip the second.
type EnvSource struct{}

// Name returns the identifier for this source.
func (e *EnvSource) Name() string { return "env" }

// Load reads all environment variables with the GENEVER_ prefix.
//
// The context is currently not used but is included for API consistency.
//
// Returns a map with nested structure based on underscore-delimited variable names.
// Never returns an error - missing or invalid environment variables are ignored.
func (e *EnvSource) Load(ctx context.Context) (map[string]any, error) {
	return loadEnvVars(), nil
}

// Watch is not implemented for EnvSource.
// Returns nil immediately, indicating that environment watching is not supported.
//
// Environment variables typically don't change during process runtime.
// If you need to detect changes, consider implementing a polling mechanism.
func (e *EnvSource) Watch(ctx context.Context, ch chan<- config.Event) error {
	return nil
}

func loadEnvVars() map[string]any {
	result := make(map[string]any)

	for _, env := range os.Environ() {
		key, value, found := parseEnvLine(env)
		if !found {
			continue
		}

		if !strings.HasPrefix(key, ENV_PREFIX) {
			continue
		}

		key = strings.TrimPrefix(key, ENV_PREFIX)
		key = strings.ToLower(key)

		segments := strings.Split(key, "_")
		if len(segments) == 0 {
			continue
		}

		setNestedValue(result, segments, value)
	}

	return result
}

func parseEnvLine(env string) (string, string, bool) {
	parts := strings.SplitN(env, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func setNestedValue(m map[string]any, segments []string, value string) {
	current := m

	for i, segment := range segments {
		if segment == "" {
			continue
		}

		if i == len(segments)-1 {
			current[segment] = value
			return
		}

		if existing, exists := current[segment]; exists {
			if nested, ok := existing.(map[string]any); ok {
				current = nested
			} else {
				// Conflict: a leaf value already exists at this path
				// Skip this entry to avoid overwriting existing data
				return
			}
		} else {
			nested := make(map[string]any)
			current[segment] = nested
			current = nested
		}
	}
}
