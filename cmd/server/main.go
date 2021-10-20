package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"grpcotel/backend"
	"grpcotel/cmd/common"
	backendrpc "grpcotel/rpc/backend"
)

func main() {
	grpcServer, tracerProvider := common.Setup("grpcotel_server")

	mux := runtime.NewServeMux()
	ctx := context.Background()
	endpoint := "localhost:8300"
	opts := []grpc.DialOption{grpc.WithInsecure()}

	backendrpc.RegisterBackendServiceServer(
		grpcServer, backend.NewServer(tracerProvider.Tracer("backend")))
	_ = backendrpc.RegisterBackendServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)

	common.StartServers(grpcServer, mux, common.ServerConfig{
		GRPC: ":8300",
		HTTP: ":8200",
	})
}
