package main

import (
	"context"
	"fmt"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"grpcotel/backend"
	backendrpc "grpcotel/rpc/backend"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger),
		),
	)

	mux := runtime.NewServeMux()
	ctx := context.Background()
	endpoint := "localhost:8300"
	opts := []grpc.DialOption{grpc.WithInsecure()}

	backendrpc.RegisterBackendServiceServer(grpcServer, &backend.Server{})
	_ = backendrpc.RegisterBackendServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)

	// Start Servers
	var wg sync.WaitGroup
	wg.Add(2)

	listener, err := net.Listen("tcp", ":8300")
	if err != nil {
		panic(err)
	}

	httpServer := &http.Server{
		Addr:    ":8200",
		Handler: mux,
	}

	go func() {
		defer wg.Done()
		fmt.Println("Start gRPC on port 8300")

		err := grpcServer.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		fmt.Println("Start HTTP on port 8200")

		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	<-exit

	grpcServer.Stop()
	err = httpServer.Shutdown(ctx)
	if err != nil {
		panic(err)
	}

	wg.Wait()
	fmt.Println("Graceful shutdown completed")
}
