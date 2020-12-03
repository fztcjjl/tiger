package client

import (
	"github.com/fztcjjl/tiger/trpc/client/resolver"
	log "github.com/fztcjjl/tiger/trpc/logger"
	"github.com/fztcjjl/tiger/trpc/registry"
	"google.golang.org/grpc"
)

type Client struct {
	options Options
	conn    *grpc.ClientConn
	r       registry.Registry
}

func NewClient(service string, opt ...Option) *Client {
	opts := newOptions(opt...)
	client := Client{options: opts}

	resolver.Register(opts.Registry)
	target := opts.Registry.String() + ":///" + service

	grpcDialOptions := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
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
