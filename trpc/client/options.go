package client

import (
	"context"
	"github.com/fztcjjl/tiger/trpc/registry"
	"github.com/fztcjjl/tiger/trpc/registry/mdns"
	"google.golang.org/grpc"
)

type Option func(*Options)

type Options struct {
	Registry    registry.Registry
	DialOptions []grpc.DialOption
	// Other opts for implementations of the interface
	// can be stored in a context
	Context context.Context
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

type unaryClientInterceptors struct{}

func Interceptors(interceptors ...grpc.UnaryClientInterceptor) Option {
	return setClientOption(unaryClientInterceptors{}, interceptors)
}
