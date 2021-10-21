package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"grpcotel/backend"
	"grpcotel/cmd/common"
	backendrpc "grpcotel/rpc/backend"
)

func main() {
	grpcServer, tracerProvider, shutdown := common.Setup("grpcotel_proxy")
	defer shutdown()

	mux := runtime.NewServeMux()
	ctx := context.Background()
	endpoint := "localhost:7300"
	opts := []grpc.DialOption{grpc.WithInsecure()}

	conn, err := grpc.Dial("localhost:8300",
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(
			otelgrpc.UnaryClientInterceptor(
				otelgrpc.WithTracerProvider(tracerProvider),
				otelgrpc.WithPropagators(propagation.TraceContext{}),
			),
		))
	if err != nil {
		panic(err)
	}

	backendrpc.RegisterBackendServiceServer(
		grpcServer, backend.NewProxy(conn, tracerProvider))
	_ = backendrpc.RegisterBackendServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)

	common.StartServers(grpcServer, mux, common.ServerConfig{
		GRPC: ":7300",
		HTTP: ":7200",
	})
}
