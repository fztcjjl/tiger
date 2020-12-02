package main

import (
	"context"
	"github.com/fztcjjl/tiger/app"
	pb "github.com/fztcjjl/tiger/examples/proto"
	"github.com/fztcjjl/tiger/trpc/client"
	"github.com/fztcjjl/tiger/trpc/registry"
	"github.com/fztcjjl/tiger/trpc/registry/etcd"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

func main() {
	a := app.NewApp(app.WithHttp(true))
	srv := a.GetServer()
	webSrv := a.GetWebServer()
	webSrv.HandleFunc("/hello", SayHello)
	pb.RegisterGreeterServer(srv.Server(), &Greeter{})
	a.Run()

}

type Greeter struct {
}

func (g *Greeter) SayHello(ctx context.Context, req *pb.HelloRequest) (rsp *pb.HelloReply, err error) {
	rsp = &pb.HelloReply{Message: "Hello " + req.Name}
	return
}

func SayHello(w http.ResponseWriter, r *http.Request) {
	cli := client.NewClient(
		"tiger.srv.hello",
		client.Registry(etcd.NewRegistry(registry.Addrs("127.0.0.1:2379"))),
		client.GrpcDialOption(grpc.WithInsecure()),
	)

	grpcClient := pb.NewGreeterClient(cli.GetConn())

	req := pb.HelloRequest{Name: "John"}
	rsp, err := grpcClient.SayHello(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Greeting: %s", rsp.Message)
	w.Write([]byte(rsp.Message))
}
