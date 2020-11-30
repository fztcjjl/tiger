package app

import (
	log "github.com/fztcjjl/tiger/trpc/logger"
	"os"
	"os/signal"
	"syscall"
)

type app struct {
	opts Options
}

func newApp(opt ...Option) App {
	app := new(app)
	options := newOptions(opt...)
	app.opts = options

	return app
}

func (a *app) Name() string {
	return a.opts.Server.Options().Name
}

func (a *app) Init(opt ...Option) {
	for _, o := range opt {
		o(&a.opts)
	}
}

func (a *app) Run() error {
	log.Infof("Starting [service] %s", a.Name())
	if err := a.opts.Server.Start(); err != nil {
		return err
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	// wait on kill signal
	case <-ch:
	// wait on context cancel
	case <-a.opts.Context.Done():
	}

	return a.opts.Server.Stop()
}
