package core

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
)

type App struct {
	Modules   []Module
	Container Container
	Logger    *slog.Logger
}

func NewApp(logger *slog.Logger, mods ...Module) *App {
	return &App{
		Modules:   mods,
		Container: NewContainer(),
		Logger:    logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	// 1) Order modules by dependencies (simple topo-sort)
	order, err := topoSort(a.Modules)
	if err != nil {
		return err
	}

	// 2) Configure
	for _, m := range order {
		if err := m.Configure(a.Container); err != nil {
			return err
		}
	}

	// 3) Start in order
	for _, m := range order {
		a.Logger.Info("starting module", "module", m.Name())
		if err := m.Start(ctx, a.Container); err != nil {
			return err
		}
	}

	// 4) Wait for signal, then stop in reverse order
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
	case <-stop:
	}

	// give modules time to shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var firstErr error
	for i := len(order) - 1; i >= 0; i-- {
		m := order[i]
		a.Logger.Info("stopping module", "module", m.Name())
		if err := m.Stop(shutdownCtx, a.Container); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func topoSort(mods []Module) ([]Module, error) {
	nameToMod := map[string]Module{}
	for _, m := range mods {
		if _, dup := nameToMod[m.Name()]; dup {
			return nil, errors.New("duplicate module name: " + m.Name())
		}
		nameToMod[m.Name()] = m
	}
	visited := map[string]bool{}
	temp := map[string]bool{}
	var out []Module
	var visit func(string) error

	visit = func(n string) error {
		if temp[n] {
			return errors.New("cycle detected at module " + n)
		}
		if visited[n] {
			return nil
		}
		temp[n] = true
		m := nameToMod[n]
		for _, d := range m.DependsOn() {
			if _, ok := nameToMod[d]; !ok {
				return errors.New("missing dependency: " + n + " depends on " + d)
			}
			if err := visit(d); err != nil {
				return err
			}
		}
		visited[n] = true
		temp[n] = false
		out = append(out, m)
		return nil
	}

	// Make iteration order stable.
	names := make([]string, 0, len(mods))
	for _, m := range mods {
		names = append(names, m.Name())
	}
	sort.Strings(names)

	for _, n := range names {
		if err := visit(n); err != nil {
			return nil, err
		}
	}
	return out, nil
}
