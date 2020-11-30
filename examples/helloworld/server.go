package main

import (
	"context"
	"github.com/fztcjjl/tiger/app"
	pb "github.com/fztcjjl/tiger/examples/proto"
	"github.com/fztcjjl/tiger/trpc/server"
)

func main() {
	srv := server.NewServer(
		server.Name("mdns.srv.hello"),
	)
	pb.RegisterGreeterServer(srv.Server(), &Greeter{})
	app := app.NewApp(app.Server(srv))
	app.Run()
}

type Greeter struct {
}

func (g *Greeter) SayHello(ctx context.Context, req *pb.HelloRequest) (rsp *pb.HelloReply, err error) {
	rsp = &pb.HelloReply{Message: "Hello " + req.Name}
	return
}
