package main

import (
	"context"
	pb "github.com/fztcjjl/tiger/examples/proto"
	"github.com/fztcjjl/tiger/trpc/registry"
	"github.com/fztcjjl/tiger/trpc/registry/etcd"
	"github.com/fztcjjl/tiger/trpc/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	srv := server.NewServer(
		server.Name("tiger.srv.hello"),
		server.Registry(etcd.NewRegistry(registry.Addrs("127.0.0.1:2379"))),
	)
	pb.RegisterGreeterServer(srv.Server(), &Greeter{})
	srv.Start()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch
	srv.Stop()

}

type Greeter struct {
}

func (g *Greeter) SayHello(ctx context.Context, req *pb.HelloRequest) (rsp *pb.HelloReply, err error) {
	log.Printf("Received: %s", req.Name)
	rsp = &pb.HelloReply{Message: "Hello " + req.Name}
	return
}
