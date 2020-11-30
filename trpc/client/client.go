package client

import (
	"github.com/fztcjjl/tiger/trpc/client/resolver"
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

	conn, err := grpc.Dial(target, opts.DialOptions...)
	if err != nil {
		return nil
	}
	client.conn = conn

	return &client
}

func (c *Client) GetConn() *grpc.ClientConn {
	return c.conn
}
