package trace

import (
	"context"
	"github.com/opentracing/opentracing-go/ext"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

const spanContextKey = "span-context"

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		var sp opentracing.Span

		spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.HTTPHeadersCarrier(c.Request.Header))
		if err != nil {
			sp = opentracing.GlobalTracer().StartSpan(
				c.Request.URL.Path,
				opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
				ext.SpanKindRPCServer,
			)
		} else {
			sp = opentracing.GlobalTracer().StartSpan(
				c.Request.URL.Path,
				opentracing.ChildOf(spanCtx),
				opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
				ext.SpanKindRPCServer,
			)
		}

		defer sp.Finish()
		md := make(map[string]string)
		if err := sp.Tracer().Inject(
			sp.Context(),
			opentracing.TextMap,
			opentracing.TextMapCarrier(md)); err != nil {
			log.Println(err)
		}

		ctx := context.Background()
		ctx = opentracing.ContextWithSpan(ctx, sp)

		c.Set(spanContextKey, ctx)
		c.Next()
	}
}

func ContextWithSpan(c *gin.Context) (ctx context.Context) {
	v, exist := c.Get(spanContextKey)
	if !exist {
		ctx = context.Background()
		return
	}

	ctx, ok := v.(context.Context)
	if !ok {
		ctx = context.Background()
	}
	return
}
