package actuator

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/skekre98/genever/config"
	"github.com/skekre98/genever/core"
	"github.com/skekre98/genever/web"
)

const Name = "actuator"

type module struct{}

func Module() core.Module { return &module{} }

func (m *module) Name() string        { return Name }
func (m *module) DependsOn() []string { return []string{web.Name} }

func (m *module) Configure(c core.Container) error {
	engine := web.Engine(c)
	cfg := core.Get[config.Root](c)

	group := engine.Group(cfg.Actuator.BasePath)

	// Health
	group.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "UP",
			"checks": []gin.H{},
		})
	})

	// Info
	group.GET("/info", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"app": gin.H{
				"name":    cfg.App.Name,
				"version": cfg.App.Version,
			},
			"runtime": gin.H{
				"go":          runtime.Version(),
				"numGoroutine": runtime.NumGoroutine(),
				"time":        time.Now().UTC().Format(time.RFC3339),
				"pid":         os.Getpid(),
			},
		})
	})

	// Metrics
	if cfg.Observability.Metrics.Enabled {
		group.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	return nil
}

func (m *module) Start(_ context.Context, _ core.Container) error { return nil }
func (m *module) Stop(_ context.Context, _ core.Container) error  { return nil }
