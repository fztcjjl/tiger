package app

import (
	"context"
	"github.com/fztcjjl/tiger/trpc/server"
	"github.com/fztcjjl/tiger/trpc/web"
)

type Options struct {
	Server    *server.Server
	WebServer *web.Server

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

func newOptions(opt ...Option) Options {
	opts := Options{
		Server:  server.DefaultServer,
		Context: context.Background(),
	}

	for _, o := range opt {
		o(&opts)
	}

	return opts
}

type Option func(*Options)

func Server(s *server.Server) Option {
	return func(o *Options) {
		o.Server = s
	}
}

func WebServer(s *web.Server) Option {
	return func(o *Options) {
		o.WebServer = s
	}
}

func Context(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}
