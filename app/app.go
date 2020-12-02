package app

import (
	log "github.com/fztcjjl/tiger/trpc/logger"
	"github.com/fztcjjl/tiger/trpc/registry"
	"github.com/fztcjjl/tiger/trpc/registry/etcd"
	"github.com/fztcjjl/tiger/trpc/server"
	"github.com/fztcjjl/tiger/trpc/web"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	opts      Options
	config    *Config
	server    *server.Server
	webServer *web.Server
}

func NewApp(opt ...Option) *App {
	app := new(App)
	app.loadConfig()
	options := newOptions(opt...)
	app.opts = options

	var r registry.Registry
	addrs := app.config.GetStringSlice("etcd")
	if len(addrs) > 0 {
		r = etcd.NewRegistry(registry.Addrs(addrs...))
	}
	if app.opts.EnableHttp {
		app.webServer = web.NewServer(
			web.Name(app.config.GetString("web.name")),
			web.Version(app.config.GetString("web.version")),
			web.Registry(r),
		)
	}

	app.server = server.NewServer(
		server.Name(app.config.GetString("rpc.name")),
		server.Version(app.config.GetString("rpc.version")),
		server.Registry(r),
	)

	return app
}

func (a *App) GetServer() *server.Server {
	return a.server
}

func (a *App) GetWebServer() *web.Server {
	return a.webServer
}

func (a *App) Name() string {
	return a.server.Options().Name
}

func (a *App) Init(opt ...Option) {
	for _, o := range opt {
		o(&a.opts)
	}
}

func (a *App) loadConfig() {
	v := viper.New()
	v.AddConfigPath("conf")
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		log.Fatal(err)
		return
	}

	a.config = &Config{Viper: v}
	return
}

func (a *App) GetConfig() *Config {
	return a.config
}

func (a *App) Run() error {
	log.Infof("Starting [service] %s", a.Name())
	if a.server != nil {
		if err := a.server.Start(); err != nil {
			return err
		}
	}

	if a.webServer != nil {
		if err := a.webServer.Start(); err != nil {
			return err
		}
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	// wait on kill signal
	case <-ch:
	// wait on context cancel
	case <-a.opts.Context.Done():
	}

	if a.server != nil {
		a.server.Stop()
	}

	if a.webServer != nil {
		a.webServer.Stop()
	}

	return nil
}
