package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/skekre98/genever/actuator"
	"github.com/skekre98/genever/config"
	"github.com/skekre98/genever/config/source"
	"github.com/skekre98/genever/core"
	"github.com/skekre98/genever/logging"
	"github.com/skekre98/genever/web"
)

func main() {
	// 1) config - loads from multiple sources with precedence:
	//    File (base + profile) -> Environment vars -> CLI flags
	var cfg config.Root

	profile := os.Getenv("APP_PROFILE")
	configPath := findConfigPath()

	sources := []config.ConfigSource{
		&source.FileSource{
			BasePath: configPath,
			Profile:  profile,
		},
		&source.EnvSource{},
		&source.CLISource{},
	}

	_, err := config.NewManager(&cfg, config.Options{AutoReload: false}, sources...)
	if err != nil {
		panic(err)
	}

	// Apply defaults
	if cfg.Actuator.BasePath == "" {
		cfg.Actuator.BasePath = "/actuator"
	}
	if cfg.Observability.Metrics.Path == "" {
		cfg.Observability.Metrics.Path = "/actuator/metrics"
	}
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":8080"
	}

	// 2) logging
	logger := logging.New().With(
		slog.String("app", cfg.App.Name),
		slog.String("version", cfg.App.Version),
	)

	// 3) compose the app
	app := core.NewApp(
		logger,
		// web server with some example routes
		web.Module(
			web.WithRoutes(func(r web.Router) {
				r.GET("/hello", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "world"})
				})
			}),
		),
		// actuator endpoints
		actuator.Module(),
	)

	// 4) seed shared objects into the container
	app.Container.Set(core.TypeKey[config.Root]{}, cfg)
	app.Container.Set(core.TypeKey[*slog.Logger]{}, logger)

	// 5) run
	if err := app.Run(context.Background()); err != nil {
		logger.Error("app error", "error", err)
		os.Exit(1)
	}
}

// findConfigPath determines the configuration directory path.
// It tries multiple locations in order of preference:
//  1. CONFIG_PATH environment variable (highest priority)
//  2. ./configs (when running from cmd/orders directory)
//  3. cmd/orders/configs (when running from project root)
//
// Returns the first path where application.yml or application.yaml exists.
func findConfigPath() string {
	// Check environment variable first
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}

	// Try common locations
	candidates := []string{
		"configs",              // Running from cmd/orders
		"cmd/orders/configs",   // Running from project root
	}

	for _, candidate := range candidates {
		if configExists(candidate) {
			return candidate
		}
	}

	// Fallback to default (will fail during load with helpful error)
	return "configs"
}

// configExists checks if a config directory has an application.yml or application.yaml file
func configExists(dir string) bool {
	for _, filename := range []string{"application.yml", "application.yaml"} {
		if _, err := os.Stat(dir + "/" + filename); err == nil {
			return true
		}
	}
	return false
}
