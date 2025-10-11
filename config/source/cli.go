package source

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/skekre98/genever/config"
	"github.com/spf13/pflag"
)

// CLISource loads configuration from command-line flags.
//
// CLISource parses flags using dot notation to create nested configuration:
//
// Flag format:
//   - Use dots to indicate nesting: --http.port=8080
//   - Supports both --flag=value and --flag value formats
//   - Supports single-dash for long flags: -http.port=8080
//   - Empty values are ignored
//   - Non-flag arguments are ignored
//
// Examples:
//
//	--server.port=8080 --server.host=localhost
//	  -> {server: {host: "localhost", port: "8080"}}
//
//	--database.host db.example.com --database.port 5432
//	  -> {database: {host: "db.example.com", port: "5432"}}
//
//	-app.name=myapp -app.debug=true
//	  -> {app: {name: "myapp", debug: "true"}}
//
// All values are returned as strings. Type conversion happens during binding.
//
// CLISource reads from os.Args and should typically be the last source in
// the precedence chain to allow command-line flags to override all other sources.
type CLISource struct{}

// Name returns the identifier for this source.
func (c *CLISource) Name() string { return "cli" }

// Load parses command-line flags from os.Args.
//
// The context is currently not used but is included for API consistency.
//
// Returns a map with nested structure based on dot-notated flag names.
// Never returns an error - invalid flags are ignored.
func (c *CLISource) Load(ctx context.Context) (map[string]any, error) {
	return parseCliFlags()
}

// Watch is not implemented for CLISource.
// Returns nil immediately, indicating that CLI argument watching is not supported.
//
// Command-line arguments are static for the process lifetime.
func (c *CLISource) Watch(ctx context.Context, ch chan<- config.Event) error {
	return nil
}

func parseCliFlags() (map[string]any, error) {
	result := make(map[string]any)
	fs := pflag.NewFlagSet("config", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)

	registeredFlags := make(map[string]bool)
	args := normalizeArgs(os.Args[1:])

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if !strings.HasPrefix(arg, "-") {
			continue
		}

		flagName := extractFlagName(arg)
		if flagName == "" {
			continue
		}

		if !registeredFlags[flagName] {
			fs.String(flagName, "", fmt.Sprintf("Config value for %s", flagName))
			registeredFlags[flagName] = true
		}

		if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			i++
		}
	}

	_ = fs.Parse(args)

	fs.VisitAll(func(flag *pflag.Flag) {
		if !flag.Changed {
			return
		}

		value := flag.Value.String()
		if value == "" {
			return
		}

		segments := strings.Split(flag.Name, ".")
		if len(segments) == 0 {
			return
		}

		setNestedValue(result, segments, value)
	})

	return result, nil
}

// normalizeArgs converts single-dash long flags to double-dash for pflag.
func normalizeArgs(args []string) []string {
	normalized := make([]string, len(args))
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			withoutDash := strings.TrimPrefix(arg, "-")
			if len(withoutDash) > 1 && withoutDash[0] != '=' {
				normalized[i] = "-" + arg
			} else {
				normalized[i] = arg
			}
		} else {
			normalized[i] = arg
		}
	}
	return normalized
}

// extractFlagName extracts the flag name, removing dashes and handling --flag=value format.
func extractFlagName(arg string) string {
	arg = strings.TrimLeft(arg, "-")
	if arg == "" {
		return ""
	}

	if idx := strings.Index(arg, "="); idx != -1 {
		return arg[:idx]
	}

	return arg
}
