package common

import (
	"context"
	"fmt"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"grpcotel/pkg/level"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

func newResource(name string) *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "local"),
		),
	)
	return r
}

func disableForHealthCheckInterceptor(interceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	const (
		healthLive  = "/backend.HealthService/Live"
		healthReady = "/backend.HealthService/Ready"
	)

	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if info.FullMethod == healthLive || info.FullMethod == healthReady {
			return handler(ctx, req)
		}
		return interceptor(ctx, req, info, handler)
	}
}

// Setup ...
func Setup(serviceName string) (*grpc.Server, trace.TracerProvider, func()) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	exporter := newJaegerExporter()
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(newResource(serviceName)),
	)
	shutdown := func() {
		err := tracerProvider.Shutdown(context.Background())
		if err != nil {
			panic(err)
		}
	}

	otelInterceptor := otelgrpc.UnaryServerInterceptor(
		otelgrpc.WithTracerProvider(tracerProvider),
		otelgrpc.WithPropagators(propagation.TraceContext{}),
	)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			disableForHealthCheckInterceptor(otelInterceptor),
			grpc_ctxtags.UnaryServerInterceptor(),
			level.SetTraceInfoInterceptor(logger),
			grpc_zap.UnaryServerInterceptor(logger),
		),
	)

	return grpcServer, tracerProvider, shutdown
}

// ServerConfig ...
type ServerConfig struct {
	GRPC string
	HTTP string
}

// StartServers ...
func StartServers(grpcServer *grpc.Server, mux *runtime.ServeMux, config ServerConfig) {
	// Start Servers
	var wg sync.WaitGroup
	wg.Add(2)

	listener, err := net.Listen("tcp", config.GRPC)
	if err != nil {
		panic(err)
	}

	httpServer := &http.Server{
		Addr:    config.HTTP,
		Handler: mux,
	}

	go func() {
		defer wg.Done()
		fmt.Println("Start gRPC on port", config.GRPC)

		err := grpcServer.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		fmt.Println("Start HTTP on port", config.HTTP)

		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	<-exit

	grpcServer.Stop()
	err = httpServer.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}

	wg.Wait()
	fmt.Println("Graceful shutdown completed")
}
