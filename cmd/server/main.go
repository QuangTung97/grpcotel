package main

import (
	"context"
	"fmt"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"grpcotel/backend"
	"grpcotel/pkg/tracing"
	backendrpc "grpcotel/rpc/backend"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/exporters/jaeger"
)

func newJaegerExporter() sdktrace.SpanExporter {
	exporter, err := jaeger.New(
		jaeger.WithAgentEndpoint(),
	)
	if err != nil {
		panic(err)
	}
	return exporter
}

func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("grpcotel"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "local"),
		),
	)
	return r
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	exporter := newJaegerExporter()

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(newResource()),
	)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(
				otelgrpc.WithTracerProvider(tracerProvider),
			),
			grpc_ctxtags.UnaryServerInterceptor(),
			tracing.SetTraceInfoInterceptor,
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
