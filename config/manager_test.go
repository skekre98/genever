package config_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/skekre98/genever/config"
)

// mockSource is a test implementation of config.ConfigSource
type mockSource struct {
	name   string
	data   map[string]any
	errVal error
	mu     sync.RWMutex
}

func (m *mockSource) Name() string {
	return m.name
}

func (m *mockSource) Load(ctx context.Context) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errVal != nil {
		return nil, m.errVal
	}

	// Return a deep copy to prevent race conditions when data is modified
	result := make(map[string]any, len(m.data))
	for k, v := range m.data {
		// For nested maps, create copies
		if nestedMap, ok := v.(map[string]any); ok {
			copiedMap := make(map[string]any, len(nestedMap))
			for nk, nv := range nestedMap {
				copiedMap[nk] = nv
			}
			result[k] = copiedMap
		} else {
			result[k] = v
		}
	}
	return result, nil
}

func (m *mockSource) Watch(ctx context.Context, ch chan<- config.Event) error {
	return nil
}

func TestNewManager_Success(t *testing.T) {
	type AppConfig struct {
		Name string `config:"name" validate:"required"`
		Port int    `config:"port" validate:"required,min=1,max=65535"`
	}

	source := &mockSource{
		name: "test",
		data: map[string]any{
			"name": "test-app",
			"port": 8080,
		},
	}

	var cfg AppConfig
	manager, err := config.NewManager(&cfg, config.Options{}, source)

	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil manager")
	}

	// Verify config was bound correctly
	if cfg.Name != "test-app" {
		t.Errorf("config.Name = %v, want %v", cfg.Name, "test-app")
	}

	if cfg.Port != 8080 {
		t.Errorf("config.Port = %v, want %v", cfg.Port, 8080)
	}
}

func TestNewManager_LoadError(t *testing.T) {
	type AppConfig struct {
		Name string `config:"name"`
	}

	source := &mockSource{
		name:   "test",
		errVal: errors.New("load error"),
	}

	var cfg AppConfig
	_, err := config.NewManager(&cfg, config.Options{}, source)

	if err == nil {
		t.Fatal("NewManager() expected error, got nil")
	}

	if !errors.Is(err, source.errVal) {
		t.Errorf("NewManager() error = %v, want to contain %v", err, source.errVal)
	}
}

func TestNewManager_ValidationError(t *testing.T) {
	type AppConfig struct {
		Name string `config:"name" validate:"required"`
		Port int    `config:"port" validate:"required,min=1,max=65535"`
	}

	source := &mockSource{
		name: "test",
		data: map[string]any{
			"name": "test-app",
			"port": 99999, // Invalid: exceeds max
		},
	}

	var cfg AppConfig
	_, err := config.NewManager(&cfg, config.Options{}, source)

	if err == nil {
		t.Fatal("NewManager() expected validation error, got nil")
	}

	var bindErr *config.BindError
	if !errors.As(err, &bindErr) {
		t.Fatalf("expected BindError, got %T", err)
	}

	if bindErr.Stage != "validate" {
		t.Errorf("expected validation error, got %s error", bindErr.Stage)
	}
}

func TestManager_Reload(t *testing.T) {
	type AppConfig struct {
		Name string `config:"name" validate:"required"`
		Port int    `config:"port" validate:"required,min=1,max=65535"`
	}

	source := &mockSource{
		name: "test",
		data: map[string]any{
			"name": "initial-app",
			"port": 8080,
		},
	}

	var cfg AppConfig
	manager, err := config.NewManager(&cfg, config.Options{}, source)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Verify initial config
	if cfg.Name != "initial-app" {
		t.Errorf("initial config.Name = %v, want %v", cfg.Name, "initial-app")
	}

	// Update source data
	source.data = map[string]any{
		"name": "updated-app",
		"port": 9090,
	}

	// Reload config
	err = manager.Reload(context.Background())
	if err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	// Verify updated config
	if cfg.Name != "updated-app" {
		t.Errorf("updated config.Name = %v, want %v", cfg.Name, "updated-app")
	}

	if cfg.Port != 9090 {
		t.Errorf("updated config.Port = %v, want %v", cfg.Port, 9090)
	}
}

func TestManager_MultipleSources(t *testing.T) {
	type DatabaseConfig struct {
		Host string `config:"host"`
		Port int    `config:"port"`
	}

	type AppConfig struct {
		Name     string         `config:"name" validate:"required"`
		Port     int            `config:"port" validate:"required"`
		Database DatabaseConfig `config:"database"`
	}

	// Base config source
	baseSource := &mockSource{
		name: "base",
		data: map[string]any{
			"name": "app",
			"port": 8080,
			"database": map[string]any{
				"host": "localhost",
				"port": 5432,
			},
		},
	}

	// Override source (should override values from base)
	overrideSource := &mockSource{
		name: "override",
		data: map[string]any{
			"port": 9090,
			"database": map[string]any{
				"host": "prod-db.example.com",
			},
		},
	}

	var cfg AppConfig
	_, err := config.NewManager(&cfg, config.Options{}, baseSource, overrideSource)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Verify merged config
	if cfg.Name != "app" {
		t.Errorf("config.Name = %v, want %v", cfg.Name, "app")
	}

	if cfg.Port != 9090 {
		t.Errorf("config.Port = %v, want %v (should be overridden)", cfg.Port, 9090)
	}

	if cfg.Database.Host != "prod-db.example.com" {
		t.Errorf("config.Database.Host = %v, want %v (should be overridden)", cfg.Database.Host, "prod-db.example.com")
	}

	// Database.Port should remain from base since override didn't set it
	if cfg.Database.Port != 5432 {
		t.Errorf("config.Database.Port = %v, want %v (should be from base)", cfg.Database.Port, 5432)
	}
}

func TestManager_Subscribe(t *testing.T) {
	type AppConfig struct {
		Name string `config:"name" validate:"required"`
	}

	source := &mockSource{
		name: "test",
		data: map[string]any{
			"name": "test-app",
		},
	}

	var cfg AppConfig
	manager, err := config.NewManager(&cfg, config.Options{}, source)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Subscribe to events
	eventCh := make(chan config.Event, 1)
	manager.Subscribe(eventCh)

	// Update source and reload
	source.data = map[string]any{
		"name": "updated-app",
	}

	err = manager.Reload(context.Background())
	if err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	// Verify event was sent
	select {
	case evt := <-eventCh:
		t.Logf("Received event: %+v", evt)
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive event, but got none")
	}
}

func TestManager_Subscribe_MultipleSubscribers(t *testing.T) {
	type AppConfig struct {
		Name string `config:"name" validate:"required"`
	}

	source := &mockSource{
		name: "test",
		data: map[string]any{
			"name": "test-app",
		},
	}

	var cfg AppConfig
	manager, err := config.NewManager(&cfg, config.Options{}, source)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Subscribe multiple channels
	ch1 := make(chan config.Event, 1)
	ch2 := make(chan config.Event, 1)
	ch3 := make(chan config.Event, 1)

	manager.Subscribe(ch1)
	manager.Subscribe(ch2)
	manager.Subscribe(ch3)

	// Trigger reload
	source.mu.Lock()
	source.data = map[string]any{
		"name": "updated-app",
	}
	source.mu.Unlock()

	err = manager.Reload(context.Background())
	if err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	// All subscribers should receive event
	timeout := time.After(100 * time.Millisecond)

	select {
	case <-ch1:
		t.Log("ch1 received event")
	case <-timeout:
		t.Error("ch1 did not receive event")
	}

	select {
	case <-ch2:
		t.Log("ch2 received event")
	case <-timeout:
		t.Error("ch2 did not receive event")
	}

	select {
	case <-ch3:
		t.Log("ch3 received event")
	case <-timeout:
		t.Error("ch3 did not receive event")
	}
}

func TestManager_ComplexConfig(t *testing.T) {
	type Server struct {
		Host    string        `config:"host" validate:"required"`
		Port    int           `config:"port" validate:"required,min=1,max=65535"`
		Timeout time.Duration `config:"timeout"`
	}

	type Database struct {
		Host     string `config:"host" validate:"required"`
		Port     int    `config:"port" validate:"required"`
		Username string `config:"username" validate:"required"`
		Password string `config:"password" validate:"required"`
	}

	type AppConfig struct {
		Name     string   `config:"name" validate:"required"`
		Version  string   `config:"version" validate:"required"`
		Server   Server   `config:"server"`
		Database Database `config:"database"`
		Features []string `config:"features"`
	}

	source := &mockSource{
		name: "test",
		data: map[string]any{
			"name":    "my-app",
			"version": "1.0.0",
			"server": map[string]any{
				"host":    "0.0.0.0",
				"port":    8080,
				"timeout": "30s",
			},
			"database": map[string]any{
				"host":     "db.example.com",
				"port":     5432,
				"username": "dbuser",
				"password": "secret",
			},
			"features": []string{"auth", "api", "metrics"},
		},
	}

	var cfg AppConfig
	manager, err := config.NewManager(&cfg, config.Options{}, source)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Verify all fields
	if cfg.Name != "my-app" {
		t.Errorf("config.Name = %v, want %v", cfg.Name, "my-app")
	}

	if cfg.Version != "1.0.0" {
		t.Errorf("config.Version = %v, want %v", cfg.Version, "1.0.0")
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("config.Server.Host = %v, want %v", cfg.Server.Host, "0.0.0.0")
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("config.Server.Port = %v, want %v", cfg.Server.Port, 8080)
	}

	if cfg.Server.Timeout != 30*time.Second {
		t.Errorf("config.Server.Timeout = %v, want %v", cfg.Server.Timeout, 30*time.Second)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("config.Database.Host = %v, want %v", cfg.Database.Host, "db.example.com")
	}

	if cfg.Database.Port != 5432 {
		t.Errorf("config.Database.Port = %v, want %v", cfg.Database.Port, 5432)
	}

	if len(cfg.Features) != 3 {
		t.Errorf("len(config.Features) = %v, want %v", len(cfg.Features), 3)
	}

	// Note: We don't test internal implementation details like manager.binder
	// in black box tests. The fact that the config was successfully bound
	// and validated is sufficient to verify the binder is working correctly.
	_ = manager
}

// TestManager_Reload_NoChangeNoNotification tests that subscribers are NOT notified
// when the config is reloaded but no values have changed
func TestManager_Reload_NoChangeNoNotification(t *testing.T) {
	type AppConfig struct {
		Name string `config:"name" validate:"required"`
		Port int    `config:"port" validate:"required"`
	}

	source := &mockSource{
		name: "test",
		data: map[string]any{
			"name": "test-app",
			"port": 8080,
		},
	}

	var cfg AppConfig
	manager, err := config.NewManager(&cfg, config.Options{}, source)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Subscribe to events
	eventCh := make(chan config.Event, 1)
	manager.Subscribe(eventCh)

	// Reload with the SAME data
	err = manager.Reload(context.Background())
	if err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	// Verify NO event was sent since config didn't change
	select {
	case evt := <-eventCh:
		t.Errorf("Expected no event when config unchanged, but received: %+v", evt)
	case <-time.After(100 * time.Millisecond):
		// This is the expected behavior - no event received
		t.Log("Correctly received no event when config unchanged")
	}
}

// TestManager_Reload_AtomicSwap tests that config updates are atomic and
// the old config is properly replaced with the new one
func TestManager_Reload_AtomicSwap(t *testing.T) {
	type AppConfig struct {
		Name string `config:"name" validate:"required"`
		Port int    `config:"port" validate:"required"`
	}

	source := &mockSource{
		name: "test",
		data: map[string]any{
			"name": "initial-app",
			"port": 8080,
		},
	}

	var cfg AppConfig
	manager, err := config.NewManager(&cfg, config.Options{}, source)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Verify initial state
	if cfg.Name != "initial-app" || cfg.Port != 8080 {
		t.Fatalf("initial config not set correctly")
	}

	// Update source
	source.mu.Lock()
	source.data = map[string]any{
		"name": "updated-app",
		"port": 9090,
	}
	source.mu.Unlock()

	// Reload
	err = manager.Reload(context.Background())
	if err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	// The config should be completely replaced with new values
	// This tests that newCfg was properly created and swapped in
	if cfg.Name != "updated-app" {
		t.Errorf("config.Name = %v, want %v (atomic swap failed)", cfg.Name, "updated-app")
	}

	if cfg.Port != 9090 {
		t.Errorf("config.Port = %v, want %v (atomic swap failed)", cfg.Port, 9090)
	}
}

// TestManager_Reload_ConcurrentSafety tests that Reload is safe to call
// concurrently from multiple goroutines
func TestManager_Reload_ConcurrentSafety(t *testing.T) {
	type AppConfig struct {
		Name    string `config:"name" validate:"required"`
		Counter int    `config:"counter"`
	}

	source := &mockSource{
		name: "test",
		data: map[string]any{
			"name":    "test-app",
			"counter": 0,
		},
	}

	var cfg AppConfig
	manager, err := config.NewManager(&cfg, config.Options{}, source)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Launch multiple goroutines that reload concurrently
	const numGoroutines = 10
	errCh := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(counter int) {
			// Update source data with proper synchronization
			source.mu.Lock()
			source.data = map[string]any{
				"name":    "test-app",
				"counter": counter,
			}
			source.mu.Unlock()

			// Reload
			if err := manager.Reload(context.Background()); err != nil {
				errCh <- err
			} else {
				errCh <- nil
			}
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("Concurrent reload %d failed: %v", i, err)
		}
	}

	// The final config should be valid (one of the counter values)
	if cfg.Name != "test-app" {
		t.Errorf("config.Name = %v, want test-app (concurrent safety failed)", cfg.Name)
	}

	// Counter should be between 0 and numGoroutines-1
	if cfg.Counter < 0 || cfg.Counter >= numGoroutines {
		t.Errorf("config.Counter = %v, want between 0 and %d (concurrent safety failed)",
			cfg.Counter, numGoroutines-1)
	}

	t.Logf("Final config after concurrent reloads: Name=%s, Counter=%d", cfg.Name, cfg.Counter)
}
