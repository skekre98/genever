package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

func Load(basePath string) (Root, error) {
	var cfg Root

	// 1) base file
	baseBytes, err := os.ReadFile(basePath)
	if err != nil {
		return cfg, fmt.Errorf("read base config: %w", err)
	}
	if err := yaml.Unmarshal(baseBytes, &cfg); err != nil {
		return cfg, fmt.Errorf("parse base config: %w", err)
	}

	// 2) profile overlay if present
	profile := os.Getenv("APP_PROFILE")
	if profile != "" {
		overlay := filepath.Join(filepath.Dir(basePath), fmt.Sprintf("application-%s.yml", profile))
		if b, err := os.ReadFile(overlay); err == nil {
			if err := yaml.Unmarshal(b, &cfg); err != nil {
				return cfg, fmt.Errorf("parse overlay %s: %w", overlay, err)
			}
		}
	}

	// 3) set defaults
	if cfg.Actuator.BasePath == "" {
		cfg.Actuator.BasePath = "/actuator"
	}
	if cfg.Observability.Metrics.Path == "" {
		cfg.Observability.Metrics.Path = "/actuator/metrics"
	}
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":8080"
	}

	return cfg, nil
}

// thin wrapper around time.ParseDuration so duration.go stays lean
func timeParseDuration(s string) (int64, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	return d.Nanoseconds(), nil
}
