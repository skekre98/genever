package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/skekre98/genever/actuator"
	"github.com/skekre98/genever/config"
	"github.com/skekre98/genever/core"
	"github.com/skekre98/genever/logging"
	"github.com/skekre98/genever/web"
)

func main() {
	// 1) config
	cfg, err := config.Load("configs/application.yml")
	if err != nil {
		panic(err)
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
