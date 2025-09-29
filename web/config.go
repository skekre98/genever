package web

import "time"

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	CertFile string `yaml:"certFile" json:"certFile"`
	KeyFile  string `yaml:"keyFile" json:"keyFile"`
}

type Server struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	BasePath     string // optional global prefix
	TLS          TLSConfig
}

type Options struct {
	// Called during Configure to register routes.
	Routes []func(r Router)
	// Optional additional middlewares.
	Middlewares []Handler
}

type Option func(*Options)

func WithRoutes(f func(r Router)) Option {
	return func(o *Options) { o.Routes = append(o.Routes, f) }
}

func WithMiddlewares(m ...Handler) Option {
	return func(o *Options) { o.Middlewares = append(o.Middlewares, m...) }
}
