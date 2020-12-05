package client

import (
	"github.com/fztcjjl/tiger/trpc/client/resolver"
	log "github.com/fztcjjl/tiger/trpc/logger"
	"github.com/fztcjjl/tiger/trpc/registry"
	"google.golang.org/grpc"
)

type Client struct {
	opts Options
	conn *grpc.ClientConn
	r    registry.Registry
}

func NewClient(service string, opt ...Option) *Client {
	opts := newOptions(opt...)
	client := Client{opts: opts}

	resolver.Register(opts.Registry)
	target := opts.Registry.String() + ":///" + service

	grpcDialOptions := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	}
	if interceptors := client.getInterceptors(); interceptors != nil {
		grpcDialOptions = append(grpcDialOptions, grpc.WithChainUnaryInterceptor(interceptors...))
	}

	grpcDialOptions = append(grpcDialOptions, opts.DialOptions...)
	conn, err := grpc.Dial(target, grpcDialOptions...)

	if err != nil {
		log.Error(err)
		return nil
	}
	client.conn = conn

	return &client
}

func (c *Client) GetConn() *grpc.ClientConn {
	return c.conn
}

func (s *Client) getInterceptors() []grpc.UnaryClientInterceptor {
	if s.opts.Context != nil {
		if v, ok := s.opts.Context.Value(unaryClientInterceptors{}).([]grpc.UnaryClientInterceptor); ok && v != nil {
			return v
		}
	}
	return nil
}
