package config

import "context"

// ConfigSource represents a source of configuration data that can be loaded
// and optionally watched for changes.
//
// Implementations include file-based sources (YAML/JSON), environment variables,
// command-line flags, remote configuration services, and more.
//
// The Load method must be safe for concurrent use. The Watch method is optional
// and may return nil if the source does not support change notifications.
type ConfigSource interface {
	// Load retrieves configuration data from this source as a string-keyed map.
	// The returned map may contain nested maps for hierarchical configuration.
	//
	// The context can be used to cancel long-running loads or enforce timeouts.
	// Implementations should check ctx.Done() and return ctx.Err() if cancelled.
	//
	// Returns an error if:
	//   - The context is cancelled (returns ctx.Err())
	//   - The source cannot be accessed, read, or parsed
	//
	// Implementations must return a copy of the data to prevent external modification.
	Load(ctx context.Context) (map[string]any, error)

	// Watch monitors this source for changes and sends events on the provided channel.
	// The channel must not be closed by the implementation.
	//
	// The context controls the lifetime of the watch operation. When the context
	// is cancelled, Watch should clean up resources and return ctx.Err().
	//
	// Returns an error if:
	//   - The context is cancelled (returns ctx.Err())
	//   - Watching is not supported (returns nil for no-op implementations)
	//   - Watching fails to initialize
	//
	// Implementations that don't support watching should return nil immediately.
	// The implementation should continue watching until the context is cancelled,
	// an error occurs, or the application terminates.
	Watch(ctx context.Context, ch chan<- Event) error

	// Name returns a human-readable identifier for this source.
	// Used in error messages, logging, and debugging.
	// Examples: "file", "env", "cli", "consul"
	Name() string
}

// Event represents a configuration change notification.
//
// Events are sent to subscribers when the configuration is reloaded and
// changes are detected. The event includes both the old and new configuration
// values, as well as a list of changed field names for fine-grained reactivity.
type Event struct {
	// ChangedKeys lists the struct field names that have different values
	// between OldConfig and NewConfig. Field names are top-level only.
	//
	// Example: If only the Server.Port changed, ChangedKeys would contain ["Server"].
	ChangedKeys []string

	// OldConfig is the configuration value before the change.
	// The actual type depends on the configuration struct passed to Manager.
	OldConfig any

	// NewConfig is the configuration value after the change.
	// The actual type depends on the configuration struct passed to Manager.
	NewConfig any
}
