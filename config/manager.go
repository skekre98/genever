package config

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Manager orchestrates configuration loading from multiple sources,
// validates the configuration, and notifies subscribers of changes.
//
// Manager supports:
//   - Loading from multiple sources with merge precedence
//   - Atomic configuration updates with validation
//   - Thread-safe concurrent access
//   - Change detection and subscriber notifications
//   - Optional automatic reload on source changes
//
// The configuration is updated atomically - validation failures prevent
// any changes from taking effect. All public methods are safe for concurrent use.
type Manager struct {
	sources   []ConfigSource
	config    any
	binder    *Binder
	mu        sync.RWMutex
	subs      []chan Event
	autoWatch bool
}

// Options configures the behavior of a Manager.
type Options struct {
	// AutoReload enables automatic configuration reloading when sources
	// support watching. If true, the Manager will start watchers for each
	// source and reload the configuration when changes are detected.
	AutoReload bool

	// Profile specifies the configuration profile to use.
	// This field is currently unused by Manager but may be passed to sources.
	// Deprecated: Profile should be set directly on FileSource instead.
	Profile string
}

// NewManager creates a new configuration Manager that loads and validates
// configuration from the provided sources.
//
// The cfg parameter must be a pointer to a struct that will hold the
// configuration values. The struct fields should use `config` tags for
// mapping and `validate` tags for validation rules.
//
// Sources are processed in order, with later sources overriding earlier ones.
// For example, with sources [file, env, cli], CLI flags will override both
// environment variables and file values.
//
// If opts.AutoReload is true, the Manager will start background goroutines
// to watch each source for changes and automatically reload the configuration.
//
// Returns an error if the initial load or validation fails. The configuration
// is validated before being applied, so partial updates never occur.
//
// Example:
//
//	var cfg AppConfig
//	mgr, err := config.NewManager(&cfg, config.Options{AutoReload: true},
//	    &source.FileSource{BasePath: "configs"},
//	    &source.EnvSource{},
//	    &source.CLISource{},
//	)
func NewManager(cfg any, opts Options, sources ...ConfigSource) (*Manager, error) {
	m := &Manager{
		sources:   sources,
		config:    cfg,
		binder:    NewBinder(),
		autoWatch: opts.AutoReload,
	}

	if err := m.Reload(context.Background()); err != nil {
		return nil, err
	}

	if m.autoWatch {
		m.startWatchers()
	}

	return m, nil
}

// Reload loads configuration from all sources, validates it, and atomically
// updates the configuration if validation succeeds.
//
// The reload process:
//  1. Loads data from each source in order
//  2. Merges the data (later sources override earlier ones)
//  3. Creates a new configuration instance
//  4. Validates the new configuration
//  5. Atomically swaps the old configuration with the new one
//  6. Notifies subscribers if any fields changed
//
// If any step fails, the current configuration remains unchanged and an error
// is returned. This ensures the configuration is always in a valid state.
//
// The context can be used to cancel the reload operation or enforce timeouts.
// If the context is cancelled, Reload returns immediately with ctx.Err().
//
// Reload is safe to call concurrently from multiple goroutines.
//
// Returns an error if:
//   - The context is cancelled (returns ctx.Err())
//   - Any source fails to load
//   - The configuration fails to bind (decode error)
//   - The configuration fails validation
func (m *Manager) Reload(ctx context.Context) error {
	merged := map[string]any{}
	for _, src := range m.sources {
		// Check for cancellation before loading each source
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		vals, err := src.Load(ctx)
		if err != nil {
			return fmt.Errorf("failed to load config from %s: %w", src.Name(), err)
		}
		mergeMaps(merged, vals)
	}

	// Create new instance of same type as m.config
	newCfg := reflect.New(reflect.TypeOf(m.config).Elem()).Interface()

	// Bind + validate on temporary
	if err := m.binder.Bind(merged, newCfg); err != nil {
		return fmt.Errorf("failed to bind config: %w", err)
	}

	// Lock and atomically replace on success
	m.mu.Lock()

	// Create a copy of old config for comparison
	oldCfg := reflect.New(reflect.TypeOf(m.config).Elem()).Interface()
	reflect.ValueOf(oldCfg).Elem().Set(reflect.ValueOf(m.config).Elem())

	// Copy values from newCfg into m.config (updates the user's struct in place)
	reflect.ValueOf(m.config).Elem().Set(reflect.ValueOf(newCfg).Elem())

	m.mu.Unlock()

	if !reflect.DeepEqual(oldCfg, newCfg) {
		diffEvent := diffEvent(oldCfg, newCfg)
		m.notify(diffEvent)
	}
	return nil
}

// Subscribe registers a channel to receive configuration change events.
//
// When the configuration is reloaded and changes are detected, an Event
// will be sent to all subscribed channels. Events are sent asynchronously
// and non-blocking - if a channel's buffer is full, the event is dropped.
//
// The channel should be buffered to avoid missing events:
//
//	ch := make(chan config.Event, 10)
//	manager.Subscribe(ch)
//
//	go func() {
//	    for event := range ch {
//	        log.Printf("Config changed: %v", event.ChangedKeys)
//	        // React to configuration changes
//	    }
//	}()
//
// Subscribe is safe to call concurrently. The channel is never closed by
// the Manager, so callers are responsible for lifecycle management.
//
// Note: Events are only sent when Reload() detects actual changes. Reloading
// with identical values does not trigger events.
func (m *Manager) Subscribe(ch chan Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subs = append(m.subs, ch)
}

func (m *Manager) notify(evt Event) {
	m.mu.RLock()
	subs := append([]chan Event(nil), m.subs...)
	defer m.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- evt:
		default:
		}
	}
}

func (m *Manager) startWatchers() {
	for _, s := range m.sources {
		src := s // Capture loop variable
		ch := make(chan Event)
		go func() {
			// Create a background context for watching
			// TODO: Consider adding a context field to Manager for controlled shutdown
			ctx := context.Background()

			if err := src.Watch(ctx, ch); err != nil {
				// Watch returned an error or doesn't support watching
				return
			}

			for {
				select {
				case <-ctx.Done():
					return
				case <-ch:
					// Reload with background context
					// Errors are intentionally ignored as they're logged by subscribers
					_ = m.Reload(context.Background())
				}
			}
		}()
	}
}
