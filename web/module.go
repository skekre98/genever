package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/skekre98/genever/config"
	"github.com/skekre98/genever/core"
)

const Name = "web"

// Keys for DI
type engineKey struct{}
type serverKey struct{}

func Engine(c core.Container) *gin.Engine {
	return core.Get[*gin.Engine](c)
}

func Module(opts ...Option) core.Module {
	var options Options
	for _, o := range opts {
		o(&options)
	}
	return &webModule{opts: options}
}

type webModule struct {
	opts   Options
	server *http.Server
}

func (m *webModule) Name() string        { return Name }
func (m *webModule) DependsOn() []string { return nil }

func (m *webModule) Configure(c core.Container) error {
	cfg := core.Get[config.Root](c)
	l := core.Get[*slog.Logger](c)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Middlewares: request ID, recovery, access log
	r.Use(RequestID())
	r.Use(RecoveryProblem(l))
	r.Use(AccessLog(l))

	// Allow other modules/app to register routes
	var root Router = r
	if cfg.Actuator.BasePath != "" && cfg.Actuator.BasePath != "/" {
		// basePath is handled by route composers in your app if desired.
	}
	for _, reg := range m.opts.Routes {
		reg(root)
	}

	// HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	core.Put[*gin.Engine](c, r)
	core.Put[*http.Server](c, srv)
	m.server = srv
	return nil
}

func (m *webModule) Start(ctx context.Context, c core.Container) error {
	cfg := core.Get[config.Root](c)
	l := core.Get[*slog.Logger](c)
	go func() {
		l.Info("http server starting", "addr", cfg.Server.Addr)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("http server error", "error", err)
		}
	}()
	return nil
}

func (m *webModule) Stop(ctx context.Context, c core.Container) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := m.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}
	return nil
}
