package client

import (
	"github.com/fztcjjl/tiger/trpc/registry"
	"github.com/fztcjjl/tiger/trpc/registry/mdns"
	"google.golang.org/grpc"
)

type Option func(*Options)

type Options struct {
	Registry    registry.Registry
	DialOptions []grpc.DialOption
}

func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

func GrpcDialOption(opt ...grpc.DialOption) Option {
	return func(o *Options) {
		o.DialOptions = opt
	}
}

func newOptions(opt ...Option) Options {
	opts := Options{}

	for _, o := range opt {
		o(&opts)
	}

	if opts.Registry == nil {
		opts.Registry = mdns.NewRegistry()
	}

	return opts
}
