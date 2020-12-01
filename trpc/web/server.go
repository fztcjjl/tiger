package web

import (
	"crypto/tls"
	"fmt"
	"github.com/fztcjjl/tiger/trpc/logger"
	"github.com/fztcjjl/tiger/trpc/registry"
	maddr "github.com/fztcjjl/tiger/trpc/util/addr"
	mnet "github.com/fztcjjl/tiger/trpc/util/net"
	mls "github.com/fztcjjl/tiger/trpc/util/tls"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Server struct {
	opts Options

	mux *http.ServeMux
	srv *registry.Service

	sync.Mutex
	running bool
	static  bool
	exit    chan chan error
}

func NewServer(opt ...Option) *Server {
	opts := newOptions(opt...)
	srv := Server{
		opts:   opts,
		mux:    http.NewServeMux(),
		static: true,
		exit:   make(chan chan error),
	}

	srv.srv = srv.genSrv()
	return &srv
}

func (s *Server) genSrv() *registry.Service {
	var host string
	var port string
	var err error

	// default host:port
	if len(s.opts.Address) > 0 {
		host, port, err = net.SplitHostPort(s.opts.Address)
		if err != nil {
			log.Fatal(err)
		}
	}

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(s.opts.Advertise) > 0 {
		host, port, err = net.SplitHostPort(s.opts.Address)
		if err != nil {
			log.Fatal(err)
		}
	}

	addr, err := maddr.Extract(host)
	if err != nil {
		log.Fatal(err)
	}

	if strings.Count(addr, ":") > 0 {
		addr = "[" + addr + "]"
	}

	return &registry.Service{
		Name:    s.opts.Name,
		Version: s.opts.Version,
		Nodes: []*registry.Node{{
			Id:      s.opts.Id,
			Address: fmt.Sprintf("%s:%s", addr, port),
		}},
	}
}

func (s *Server) run() {
	if s.opts.RegisterInterval <= time.Duration(0) {
		return
	}

	t := time.NewTicker(s.opts.RegisterInterval)
	// return error chan
	var ch chan error

Loop:
	for {
		select {
		case <-t.C:
			s.register()
		case ch = <-s.exit:
			t.Stop()
			break Loop
		}
	}
	// deregister self
	if err := s.deregister(); err != nil {
		log.Error("Server deregister error: ", err)
	}

	ch <- nil
}

func (s *Server) register() error {
	if s.srv == nil {
		return nil
	}

	r := s.opts.Registry

	// service node need modify, node address maybe changed
	srv := s.genSrv()
	srv.Endpoints = s.srv.Endpoints
	s.srv = srv

	return r.Register(s.srv, registry.RegisterTTL(s.opts.RegisterTTL))
}

func (s *Server) deregister() error {
	if s.srv == nil {
		return nil
	}
	r := s.opts.Registry

	return r.Deregister(s.srv)
}

func (s *Server) start() error {
	s.Lock()
	defer s.Unlock()

	if s.running {
		return nil
	}

	l, err := s.listen("tcp", s.opts.Address)
	if err != nil {
		return err
	}

	s.opts.Address = l.Addr().String()
	srv := s.genSrv()
	srv.Endpoints = s.srv.Endpoints
	s.srv = srv

	var h http.Handler

	if s.opts.Handler != nil {
		h = s.opts.Handler
	} else {
		h = s.mux
		var r sync.Once

		// register the html dir
		r.Do(func() {
			// static dir
			static := s.opts.StaticDir
			if s.opts.StaticDir[0] != '/' {
				dir, _ := os.Getwd()
				static = filepath.Join(dir, static)
			}

			// set static if no / handler is registered
			if s.static {
				_, err := os.Stat(static)
				if err == nil {
					if logger.V(logger.InfoLevel, log) {
						log.Infof("Enabling static file serving from %s", static)
					}
					s.mux.Handle("/", http.FileServer(http.Dir(static)))
				}
			}
		})
	}

	var httpSrv *http.Server
	if s.opts.Server != nil {
		httpSrv = s.opts.Server
	} else {
		httpSrv = &http.Server{}
	}

	httpSrv.Handler = h

	go httpSrv.Serve(l)

	s.running = true

	log.Infof("Listening on %v", l.Addr().String())
	return nil
}

func (s *Server) Stop() error {
	s.Lock()
	defer s.Unlock()

	if !s.running {
		return nil
	}

	ch := make(chan error)
	s.exit <- ch
	var err error
	select {
	case err = <-ch:
		s.running = false
	}

	if logger.V(logger.InfoLevel, log) {
		log.Info("Stopping")
	}

	return err
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	var seen bool
	for _, ep := range s.srv.Endpoints {
		if ep.Name == pattern {
			seen = true
			break
		}
	}

	// if its unseen then add an endpoint
	if !seen {
		s.srv.Endpoints = append(s.srv.Endpoints, &registry.Endpoint{
			Name: pattern,
		})
	}

	// disable static serving
	if pattern == "/" {
		s.Lock()
		s.static = false
		s.Unlock()
	}

	// register the handler
	s.mux.Handle(pattern, handler)
}

func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	var seen bool
	for _, ep := range s.srv.Endpoints {
		if ep.Name == pattern {
			seen = true
			break
		}
	}
	if !seen {
		s.srv.Endpoints = append(s.srv.Endpoints, &registry.Endpoint{
			Name: pattern,
		})
	}

	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) Init(opts ...Option) error {
	for _, o := range opts {
		o(&s.opts)
	}

	srv := s.genSrv()
	srv.Endpoints = s.srv.Endpoints
	s.srv = srv

	return nil
}

func (s *Server) Start() error {
	if err := s.start(); err != nil {
		return err
	}

	if err := s.register(); err != nil {
		return err
	}

	// start reg loop
	go s.run()

	return nil
}

// Options returns the options for the given service
func (s *Server) Options() Options {
	return s.opts
}

func (s *Server) listen(network, addr string) (net.Listener, error) {
	var l net.Listener
	var err error

	// TODO: support use of listen options
	if s.opts.Secure || s.opts.TLSConfig != nil {
		config := s.opts.TLSConfig

		fn := func(addr string) (net.Listener, error) {
			if config == nil {
				hosts := []string{addr}

				// check if its a valid host:port
				if host, _, err := net.SplitHostPort(addr); err == nil {
					if len(host) == 0 {
						hosts = maddr.IPs()
					} else {
						hosts = []string{host}
					}
				}

				// generate a certificate
				cert, err := mls.Certificate(hosts...)
				if err != nil {
					return nil, err
				}
				config = &tls.Config{Certificates: []tls.Certificate{cert}}
			}
			return tls.Listen(network, addr, config)
		}

		l, err = mnet.Listen(addr, fn)
	} else {
		fn := func(addr string) (net.Listener, error) {
			return net.Listen(network, addr)
		}

		l, err = mnet.Listen(addr, fn)
	}

	if err != nil {
		return nil, err
	}

	return l, nil
}
