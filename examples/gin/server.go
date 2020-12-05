package main

import (
	"context"
	"github.com/fztcjjl/tiger/app"
	pb "github.com/fztcjjl/tiger/examples/proto"
	"github.com/fztcjjl/tiger/pkg/middleware/gin/trace"
	"github.com/fztcjjl/tiger/trpc/client"
	"github.com/fztcjjl/tiger/trpc/registry"
	"github.com/fztcjjl/tiger/trpc/registry/etcd"
	"github.com/fztcjjl/tiger/trpc/web"
	"github.com/gin-gonic/gin"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

func main() {
	a := app.NewApp(app.WithHttp(true))
	srv := a.GetServer()
	webSrv := a.GetWebServer()
	webSrv.Init(web.Handler(handler()))
	pb.RegisterGreeterServer(srv.Server(), &Greeter{})
	a.Run()

}

type Greeter struct {
}

func (g *Greeter) SayHello(ctx context.Context, req *pb.HelloRequest) (rsp *pb.HelloReply, err error) {
	rsp = &pb.HelloReply{Message: "Hello " + req.Name}
	return
}

func handler() http.Handler {
	route := gin.New()

	route.Use(trace.Trace())
	route.GET("/hello", sayHello)
	return route
}

func sayHello(ctx *gin.Context) {
	cli := client.NewClient(
		"srv.hello",
		client.Registry(etcd.NewRegistry(registry.Addrs("127.0.0.1:2379"))),
		client.GrpcDialOption(grpc.WithInsecure()),
		client.Interceptors(
			grpc_opentracing.UnaryClientInterceptor(),
		),
	)

	grpcClient := pb.NewGreeterClient(cli.GetConn())
	req := pb.HelloRequest{Name: "John"}
	rsp, err := grpcClient.SayHello(trace.ContextWithSpan(ctx), &req)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Greeting: %s", rsp.Message)
	ctx.Writer.WriteString(rsp.Message)
}
