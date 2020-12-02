package app

import (
	"context"
)

type Options struct {
	//Server     *server.Server
	//WebServer  *web.Server
	EnableHttp bool

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

func newOptions(opt ...Option) Options {
	opts := Options{
		//Server:  server.DefaultServer,
		Context: context.Background(),
	}

	for _, o := range opt {
		o(&opts)
	}

	return opts
}

type Option func(*Options)

func Context(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

func WithHttp(enable bool) Option {
	return func(o *Options) {
		o.EnableHttp = enable
	}
}
