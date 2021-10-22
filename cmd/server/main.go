package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"grpcotel/backend"
	"grpcotel/cmd/common"
	backendrpc "grpcotel/rpc/backend"
	"net/textproto"
)

func traceContextHeaderMatcher(key string) (string, bool) {
	const (
		traceparentHeader = "Traceparent"
		tracestateHeader  = "Tracestate"
	)

	key = textproto.CanonicalMIMEHeaderKey(key)
	switch key {
	case traceparentHeader, tracestateHeader:
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func main() {
	grpcServer, tracerProvider, shutdown := common.Setup("grpcotel_server")
	defer shutdown()

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(traceContextHeaderMatcher),
	)
	ctx := context.Background()
	endpoint := "localhost:8300"
	opts := []grpc.DialOption{grpc.WithInsecure()}

	backendrpc.RegisterHealthServiceServer(grpcServer, &backend.HealthServer{})
	backendrpc.RegisterBackendServiceServer(
		grpcServer, backend.NewServer(tracerProvider.Tracer("backend.server")))

	_ = backendrpc.RegisterHealthServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	_ = backendrpc.RegisterBackendServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)

	common.StartServers(grpcServer, mux, common.ServerConfig{
		GRPC: ":8300",
		HTTP: ":8200",
	})
}
