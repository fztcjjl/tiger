module github.com/fztcjjl/tiger

go 1.15

replace (
	go.etcd.io/etcd/api/v3 => go.etcd.io/etcd/api/v3 v3.0.0-20210204162551-dae29bb719dd
	go.etcd.io/etcd/pkg/v3 => go.etcd.io/etcd/pkg/v3 v3.0.0-20210204162551-dae29bb719dd
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/gin-gonic/gin v1.6.3
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/miekg/dns v1.1.38
	github.com/mitchellh/hashstructure v1.1.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/spf13/viper v1.7.1
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible
	go.etcd.io/etcd/api/v3 v3.5.0-pre
	go.etcd.io/etcd/client/v3 v3.0.0-20210211191245-3ef8a2d30adb
	go.uber.org/zap v1.16.0
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.25.0
)
