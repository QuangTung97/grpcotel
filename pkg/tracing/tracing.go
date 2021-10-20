package tracing

import (
	"context"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// SetTraceInfoInterceptor ...
func SetTraceInfoInterceptor(
	ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	tags := grpc_ctxtags.Extract(ctx)
	sc := trace.SpanContextFromContext(ctx)

	tags.Set("trace.id", sc.TraceID())
	tags.Set("span.id", sc.SpanID())
	tags.Set("trace.flags", sc.TraceFlags())

	return handler(ctx, req)
}
