package config

import "gopkg.in/yaml.v3"

type AppInfo struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Path    string `yaml:"path" json:"path"` // default /actuator/metrics
}

type ObservabilityConfig struct {
	Metrics MetricsConfig `yaml:"metrics" json:"metrics"`
}

type ActuatorConfig struct {
	BasePath string `yaml:"basePath" json:"basePath"` // default /actuator
}

type ServerConfig struct {
	Addr         string `yaml:"addr" json:"addr"`
	ReadTimeout  Duration `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout Duration `yaml:"writeTimeout" json:"writeTimeout"`
	IdleTimeout  Duration `yaml:"idleTimeout" json:"idleTimeout"`
}

type Root struct {
	App           AppInfo             `yaml:"app" json:"app"`
	Server        ServerConfig        `yaml:"server" json:"server"`
	Observability ObservabilityConfig `yaml:"observability" json:"observability"`
	Actuator      ActuatorConfig      `yaml:"actuator" json:"actuator"`
	// Add security, database, etc. as you grow.
}

// Duration is a thin wrapper to allow "5s" YAML parsing.
type Duration struct{ Value int64 } // nanoseconds

func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err == nil {
		// parse "5s", "1m", etc.
		if dur, err := parseDuration(s); err == nil {
			d.Value = dur
			return nil
		} else {
			return err
		}
	}
	// Fallback: parse as integer nanoseconds
	var n int64
	if err := node.Decode(&n); err != nil {
		return err
	}
	d.Value = n
	return nil
}

func parseDuration(s string) (int64, error) {
	// delegate to time.ParseDuration
	type td interface{ Parse(string) (int64, error) }
	return timeParseDuration(s)
}

// implemented separately to avoid importing time in this file.
