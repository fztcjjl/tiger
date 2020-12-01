package server

import (
	"crypto/tls"
	log "github.com/fztcjjl/tiger/trpc/logger"
	"github.com/fztcjjl/tiger/trpc/registry"
	"github.com/fztcjjl/tiger/trpc/util/addr"
	"github.com/fztcjjl/tiger/trpc/util/backoff"
	mnet "github.com/fztcjjl/tiger/trpc/util/net"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	// DefaultMaxMsgSize define maximum message size that server can send
	// or receive.  Default value is 4MB.
	DefaultMaxMsgSize = 1024 * 1024 * 4
)

const (
	defaultContentType = "application/grpc"
)

type Server struct {
	server *grpc.Server
	exit   chan chan error
	//wg     *sync.WaitGroup
	sync.RWMutex
	opts       Options
	started    bool
	registered bool
}

func NewServer(opt ...Option) *Server {
	opts := newOptions(opt...)
	srv := Server{
		opts: opts,
		exit: make(chan chan error),
	}
	srv.configure()
	return &srv
}

func (s *Server) Options() Options {
	return s.opts
}

func (s *Server) configure(opts ...Option) {
	s.Lock()
	defer s.Unlock()

	if len(opts) == 0 && s.server != nil {
		return
	}

	for _, o := range opts {
		o(&s.opts)
	}

	//s.wg = wait(s.opts.Context)

	maxMsgSize := s.getMaxMsgSize()

	gopts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	}

	if creds := s.getCredentials(); creds != nil {
		gopts = append(gopts, grpc.Creds(creds))
	}

	if opts := s.getGrpcOptions(); opts != nil {
		gopts = append(gopts, opts...)
	}
	s.server = grpc.NewServer(gopts...)

}

func (s *Server) getCredentials() credentials.TransportCredentials {
	if s.opts.Context != nil {
		if v, ok := s.opts.Context.Value(tlsAuth{}).(*tls.Config); ok && v != nil {
			return credentials.NewTLS(v)
		}
	}
	return nil
}

func (s *Server) getGrpcOptions() []grpc.ServerOption {
	if s.opts.Context == nil {
		return nil
	}

	opts, ok := s.opts.Context.Value(grpcOptions{}).([]grpc.ServerOption)
	if !ok || opts == nil {
		return nil
	}

	return opts
}

func (s *Server) getListener() net.Listener {
	if s.opts.Context == nil {
		return nil
	}

	if l, ok := s.opts.Context.Value(netListener{}).(net.Listener); ok && l != nil {
		return l
	}

	return nil
}

func (s *Server) getMaxMsgSize() int {
	if s.opts.Context == nil {
		return DefaultMaxMsgSize
	}
	size, ok := s.opts.Context.Value(maxMsgSizeKey{}).(int)
	if !ok {
		return DefaultMaxMsgSize
	}
	return size
}

func (s *Server) Start() error {
	s.RLock()
	if s.started {
		s.RUnlock()
		return nil
	}
	s.RUnlock()

	config := s.Options()

	var ts net.Listener

	if l := s.getListener(); l != nil {
		ts = l
	} else {
		var err error

		// check the tls config for secure connect
		if tc := config.TLSConfig; tc != nil {
			ts, err = tls.Listen("tcp", config.Address, tc)
			// otherwise just plain tcp listener
		} else {
			ts, err = net.Listen("tcp", config.Address)
		}
		if err != nil {
			return err
		}
	}

	if s.opts.Context != nil {
		if c, ok := s.opts.Context.Value(maxConnKey{}).(int); ok && c > 0 {
			ts = netutil.LimitListener(ts, c)
		}
	}

	log.Infof("Server [grpc] Listening on %s", ts.Addr().String())

	s.Lock()
	s.opts.Address = ts.Addr().String()
	s.Unlock()

	if err := s.register(); err != nil {
		log.Errorf("Server register error: %v", err)
	}

	go func() {
		if err := s.server.Serve(ts); err != nil {
			log.Errorf("gRPC Server start error: %v", err)
		}
	}()

	go func() {
		t := new(time.Ticker)

		// only process if it exists
		if s.opts.RegisterInterval > time.Duration(0) {
			// new ticker
			t = time.NewTicker(s.opts.RegisterInterval)
		}

		// return error chan
		var ch chan error

	Loop:
		for {
			select {
			// register self on interval
			case <-t.C:
				if err := s.register(); err != nil {
					log.Error("Server register error: ", err)
				}
			// wait for exit
			case ch = <-s.exit:
				break Loop
			}
		}

		// deregister self
		if err := s.deregister(); err != nil {
			log.Error("Server deregister error: ", err)
		}

		// wait for waitgroup
		//if s.wg != nil {
		//	s.wg.Wait()
		//}

		// stop the grpc server
		exit := make(chan bool)

		go func() {
			s.server.GracefulStop()
			close(exit)
		}()

		select {
		case <-exit:
		case <-time.After(time.Second):
			s.server.Stop()
		}

		ch <- nil
	}()

	s.Lock()
	s.started = true
	s.Unlock()
	return nil
}

func (s *Server) Stop() error {
	s.RLock()
	if !s.started {
		s.RUnlock()
		return nil
	}
	s.RUnlock()

	ch := make(chan error)
	s.exit <- ch

	var err error
	select {
	case err = <-ch:
		s.Lock()
		s.started = false
		s.Unlock()
	}

	return err
}

func (s *Server) Server() *grpc.Server {
	return s.server
}

func (s *Server) String() string {
	return "grpc"
}

func (s *Server) Init(opt ...Option) error {
	s.configure(opt...)
	return nil
}

func (s *Server) register() error {
	s.RLock()
	config := s.opts
	s.RUnlock()

	// only register if it exists or is not noop
	if config.Registry == nil || config.Registry.String() == "noop" {
		return nil
	}

	regFunc := func(service *registry.Service) error {
		var regErr error

		for i := 0; i < 3; i++ {
			// set the ttl and namespace
			rOpts := []registry.RegisterOption{
				registry.RegisterTTL(config.RegisterTTL),
				registry.RegisterDomain(s.opts.Namespace),
			}

			// attempt to register
			if err := config.Registry.Register(service, rOpts...); err != nil {
				// set the error
				regErr = err
				// backoff then retry
				time.Sleep(backoff.Do(i + 1))
				continue
			}
			// success so nil error
			regErr = nil
			break
		}

		return regErr
	}

	var err error
	var advt, host, port string

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(config.Advertise) > 0 {
		advt = config.Advertise
	} else {
		advt = config.Address
	}

	if cnt := strings.Count(advt, ":"); cnt >= 1 {
		// ipv6 address in format [host]:port or ipv4 host:port
		host, port, err = net.SplitHostPort(advt)
		if err != nil {
			return err
		}
	} else {
		host = advt
	}
	addr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	// register service
	node := &registry.Node{
		Id:      config.Name + "-" + config.Id,
		Address: mnet.HostPort(addr, port),
	}

	svc := &registry.Service{
		Name:    s.opts.Name,
		Version: s.opts.Version,
		Nodes:   []*registry.Node{node},
	}

	s.RLock()
	registered := s.registered
	s.RUnlock()

	if !registered {
		log.Infof("Registry [%s] Registering node: %s", config.Registry.String(), node.Id)
	}

	if err = regFunc(svc); err != nil {
		return err
	}

	s.registered = true

	return err
}

func (s *Server) deregister() error {
	var err error
	var advt, host, port string

	s.RLock()
	config := s.opts
	s.RUnlock()

	// only register if it exists or is not noop
	if config.Registry == nil || config.Registry.String() == "noop" {
		return nil
	}

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(config.Advertise) > 0 {
		advt = config.Advertise
	} else {
		advt = config.Address
	}

	if cnt := strings.Count(config.Address, ":"); cnt >= 1 {
		// ipv6 address in format [host]:port or ipv4 host:port
		host, port, err = net.SplitHostPort(advt)
		if err != nil {
			return err
		}
	} else {
		host = advt
	}

	addr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	node := &registry.Node{
		Id:      config.Name + "-" + config.Id,
		Address: mnet.HostPort(addr, port),
	}

	service := &registry.Service{
		Name:    config.Name,
		Version: config.Version,
		Nodes:   []*registry.Node{node},
	}

	log.Infof("Deregistering node: %s", node.Id)

	opt := registry.DeregisterDomain(s.opts.Namespace)
	if err := config.Registry.Deregister(service, opt); err != nil {
		return err
	}

	s.Lock()
	s.registered = false
	s.Unlock()
	return nil
}
