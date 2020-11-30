package server

import (
	"context"
	"crypto/tls"
	"github.com/fztcjjl/tiger/trpc/registry"
	"github.com/fztcjjl/tiger/trpc/registry/mdns"
	"github.com/fztcjjl/tiger/trpc/util/uuid"
	"google.golang.org/grpc"
	"net"
	"time"
)

type Option func(*Options)

var (
	DefaultAddress          = ":0"
	DefaultName             = "tiger.server"
	DefaultVersion          = "latest"
	DefaultId               = uuid.New().String()
	DefaultRegisterInterval = time.Second * 30
	DefaultRegisterTTL      = time.Second * 90
	DefaultServer           = NewServer()
)

type Options struct {
	Registry         registry.Registry
	Name             string
	Address          string
	Advertise        string
	Id               string
	Namespace        string
	Version          string
	RegisterTTL      time.Duration
	RegisterInterval time.Duration

	// TLSConfig specifies tls.Config for secure serving
	TLSConfig *tls.Config

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

func newOptions(opt ...Option) Options {
	opts := Options{
		RegisterTTL:      DefaultRegisterTTL,
		RegisterInterval: DefaultRegisterInterval,
	}

	for _, o := range opt {
		o(&opts)
	}

	if opts.Registry == nil {
		opts.Registry = mdns.NewRegistry()
	}

	if opts.Id == "" {
		opts.Id = DefaultId
	}

	if opts.Address == "" {
		opts.Address = DefaultAddress
	}

	if opts.Name == "" {
		opts.Name = DefaultName
	}

	if opts.Version == "" {
		opts.Version = DefaultVersion
	}

	return opts
}

// Server name
func Name(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

// Namespace to register handlers in
func Namespace(n string) Option {
	return func(o *Options) {
		o.Namespace = n
	}
}

// Unique server id
func Id(id string) Option {
	return func(o *Options) {
		o.Id = id
	}
}

// Version of the service
func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

// Address to bind to - host:port
func Address(a string) Option {
	return func(o *Options) {
		o.Address = a
	}
}

// The address to advertise for discovery - host:port
func Advertise(a string) Option {
	return func(o *Options) {
		o.Advertise = a
	}
}

// Context specifies a context for the service.
// Can be used to signal shutdown of the service
// Can be used for extra option values.
func Context(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

// Registry used for discovery
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

// Register the service with a TTL
func RegisterTTL(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterTTL = t
	}
}

// Register the service with at interval
func RegisterInterval(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterInterval = t
	}
}

// Wait tells the server to wait for requests to finish before exiting
// If `wg` is nil, server only wait for completion of rpc handler.
// For user need finer grained control, pass a concrete `wg` here, server will
// wait against it on stop.
//func Wait(wg *sync.WaitGroup) Option {
//	return func(o *Options) {
//		if o.Context == nil {
//			o.Context = context.Background()
//		}
//		if wg == nil {
//			wg = new(sync.WaitGroup)
//		}
//		o.Context = context.WithValue(o.Context, "wait", wg)
//	}
//}

type grpcOptions struct{}
type netListener struct{}
type maxMsgSizeKey struct{}
type maxConnKey struct{}
type tlsAuth struct{}

// AuthTLS should be used to setup a secure authentication using TLS
func AuthTLS(t *tls.Config) Option {
	return setServerOption(tlsAuth{}, t)
}

// MaxConn specifies maximum number of max simultaneous connections to server
func MaxConn(n int) Option {
	return setServerOption(maxConnKey{}, n)
}

// Listener specifies the net.Listener to use instead of the default
func Listener(l net.Listener) Option {
	return setServerOption(netListener{}, l)
}

// GrpcOptions to be used to configure gRPC options
func GrpcOptions(opts ...grpc.ServerOption) Option {
	return setServerOption(grpcOptions{}, opts)
}

//
// MaxMsgSize set the maximum message in bytes the server can receive and
// send.  Default maximum message size is 4 MB.
//
func MaxMsgSize(s int) Option {
	return setServerOption(maxMsgSizeKey{}, s)
}
