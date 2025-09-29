package core

import "context"

// Module is a unit of capability that participates in the app lifecycle.
type Module interface {
	Name() string
	// DependsOn declares hard dependencies by module name.
	DependsOn() []string
	// Configure registers objects into the container.
	Configure(c Container) error
	// Start begins any long-running work or servers.
	Start(ctx context.Context, c Container) error
	// Stop gracefully stops the module.
	Stop(ctx context.Context, c Container) error
}
