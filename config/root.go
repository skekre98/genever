package config

import "time"

type AppInfo struct {
	Name    string `config:"name" validate:"required"`
	Version string `config:"version" validate:"required"`
}

type MetricsConfig struct {
	Enabled bool   `config:"enabled"`
	Path    string `config:"path"`
}

type ObservabilityConfig struct {
	Metrics MetricsConfig `config:"metrics"`
}

type ActuatorConfig struct {
	BasePath string `config:"basePath"`
}

type ServerConfig struct {
	Addr         string        `config:"addr" validate:"required"`
	ReadTimeout  time.Duration `config:"readTimeout"`
	WriteTimeout time.Duration `config:"writeTimeout"`
	IdleTimeout  time.Duration `config:"idleTimeout"`
}

type Root struct {
	App           AppInfo             `config:"app"`
	Server        ServerConfig        `config:"server"`
	Observability ObservabilityConfig `config:"observability"`
	Actuator      ActuatorConfig      `config:"actuator"`
}
