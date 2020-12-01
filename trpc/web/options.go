package web

import (
	"context"
	"crypto/tls"
	"github.com/fztcjjl/tiger/trpc/logger"
	"github.com/fztcjjl/tiger/trpc/registry"
	"github.com/fztcjjl/tiger/trpc/registry/mdns"
	"github.com/fztcjjl/tiger/trpc/util/uuid"
	"net/http"
	"time"
)

type Option func(*Options)

type Options struct {
	Name      string
	Version   string
	Id        string
	Metadata  map[string]string
	Address   string
	Advertise string

	RegisterTTL      time.Duration
	RegisterInterval time.Duration

	Server  *http.Server
	Handler http.Handler

	// Alternative Options
	Context context.Context

	Registry registry.Registry

	Secure    bool
	TLSConfig *tls.Config

	// Static directory
	StaticDir string
}

func newOptions(opt ...Option) Options {
	opts := Options{
		Name:             DefaultName,
		Version:          DefaultVersion,
		Id:               DefaultId,
		Address:          DefaultAddress,
		RegisterTTL:      DefaultRegisterTTL,
		RegisterInterval: DefaultRegisterInterval,
		StaticDir:        DefaultStaticDir,
		Context:          context.TODO(),
	}

	for _, o := range opt {
		o(&opts)
	}

	if opts.Registry == nil {
		opts.Registry = mdns.NewRegistry()
	}

	return opts
}

var (
	// For serving
	DefaultName    = "tiger-web"
	DefaultVersion = "latest"
	DefaultId      = uuid.New().String()
	DefaultAddress = ":0"

	// for registration
	DefaultRegisterTTL      = time.Minute
	DefaultRegisterInterval = time.Second * 30

	// static directory
	DefaultStaticDir = "html"
	//DefaultRegisterCheck = func(context.Context) error { return nil }

	log = logger.NewHelper(logger.DefaultLogger).WithFields(map[string]interface{}{"service": "web"})
)

// Name of Web
func Name(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

// Icon specifies an icon url to load in the UI
func Icon(ico string) Option {
	return func(o *Options) {
		if o.Metadata == nil {
			o.Metadata = make(map[string]string)
		}
		o.Metadata["icon"] = ico
	}
}

//Id for Unique server id
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

// Metadata associated with the service
func Metadata(md map[string]string) Option {
	return func(o *Options) {
		o.Metadata = md
	}
}

// Address to bind to - host:port
func Address(a string) Option {
	return func(o *Options) {
		o.Address = a
	}
}

//Advertise The address to advertise for discovery - host:port
func Advertise(a string) Option {
	return func(o *Options) {
		o.Advertise = a
	}
}

// Context specifies a context for the service.
// Can be used to signal shutdown of the service.
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

//RegisterTTL register the service with a TTL
func RegisterTTL(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterTTL = t
	}
}

//RegisterInterval register the service with at interval
func RegisterInterval(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterInterval = t
	}
}

//Handler for custom handler
func Handler(h http.Handler) Option {
	return func(o *Options) {
		o.Handler = h
	}
}

// Secure Use secure communication. If TLSConfig is not specified we use InsecureSkipVerify and generate a self signed cert
func Secure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

// TLSConfig to be used for the transport.
func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

// StaticDir sets the static file directory. This defaults to ./html
func StaticDir(d string) Option {
	return func(o *Options) {
		o.StaticDir = d
	}
}

//Server for custom Server
func HttpServer(srv *http.Server) Option {
	return func(o *Options) {
		o.Server = srv
	}
}
