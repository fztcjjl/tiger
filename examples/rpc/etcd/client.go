package main

import (
	"context"
	pb "github.com/fztcjjl/tiger/examples/proto"
	"github.com/fztcjjl/tiger/trpc/client"
	"github.com/fztcjjl/tiger/trpc/registry"
	"github.com/fztcjjl/tiger/trpc/registry/etcd"
	"google.golang.org/grpc"
	"log"
)

func main() {
	cli := client.NewClient(
		"tiger.srv.hello",
		client.Registry(etcd.NewRegistry(registry.Addrs("127.0.0.1:2379"))),
		client.GrpcDialOption(grpc.WithInsecure()),
	)

	defer cli.GetConn().Close()

	grpcClient := pb.NewGreeterClient(cli.GetConn())

	req := pb.HelloRequest{Name: "John"}
	rsp, err := grpcClient.SayHello(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Greeting: %s", rsp.Message)
}
