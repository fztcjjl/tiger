package resolver

import (
	"context"
	"github.com/fztcjjl/tiger/trpc/registry"
	"google.golang.org/grpc/resolver"
	"time"
)

func Register(r registry.Registry) {
	resolver.Register(&trpcResolverBuilder{registry: r})
}

type trpcResolverBuilder struct {
	registry registry.Registry
}

func (b *trpcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &trpcResolver{
		target: target,
		cc:     cc,
		ctx:    ctx,
		cancel: cancel,
		r:      b.registry,
	}

	go r.watch()

	return r, nil
}

func (b *trpcResolverBuilder) Scheme() string {
	return b.registry.String()
}

type trpcResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
	ctx    context.Context
	cancel context.CancelFunc
	r      registry.Registry
}

func (r *trpcResolver) watch() {
	r.update()
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-r.ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			r.update()
		}
	}
}

func (*trpcResolver) ResolveNow(o resolver.ResolveNowOptions) {

}

func (r *trpcResolver) Close() {
	r.cancel()
}

func (r *trpcResolver) update() {
	var addrs []resolver.Address
	svcs, _ := r.r.GetService(r.target.Endpoint)
	for _, svc := range svcs {
		for _, node := range svc.Nodes {
			addr := resolver.Address{Addr: node.Address}
			addrs = append(addrs, addr)
		}
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}
