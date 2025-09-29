package core

import (
	"fmt"
	"reflect"
	"sync"
)

// A tiny, type-safe-ish container for the skeleton.
// (TODO: swap this out later for dig/fx or codegen wiring.)
type Container interface {
	Set(key any, val any)
	Get(key any) (any, bool)
	MustGet(key any) any
}

type container struct {
	mu  sync.RWMutex
	reg map[any]any
}

func NewContainer() Container {
	return &container{reg: make(map[any]any)}
}

func (c *container) Set(key, val any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reg[key] = val
}

func (c *container) Get(key any) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.reg[key]
	return v, ok
}

func (c *container) MustGet(key any) any {
	if v, ok := c.Get(key); ok {
		return v
	}
	panic(fmt.Errorf("container: missing dependency %v (%T)", key, key))
}

// Helpers for typed keys
type TypeKey[T any] struct{}

func Put[T any](c Container, v T) { c.Set(TypeKey[T]{}, v) }

func Get[T any](c Container) T {
	raw := c.MustGet(TypeKey[T]{})
	v, ok := raw.(T)
	if !ok {
		panic(fmt.Errorf("container: wrong type. have=%T want=%v", raw, reflect.TypeFor[T]()))
	}
	return v
}
