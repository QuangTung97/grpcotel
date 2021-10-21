package backend

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	backendrpc "grpcotel/rpc/backend"
)

// Proxy ...
type Proxy struct {
	backendrpc.UnimplementedBackendServiceServer
	tracer trace.Tracer
	conn   *grpc.ClientConn
}

// NewProxy ...
func NewProxy(conn *grpc.ClientConn, tp trace.TracerProvider) *Proxy {
	return &Proxy{tracer: tp.Tracer("backend.proxy"), conn: conn}
}

// GetUser ...
func (p *Proxy) GetUser(
	ctx context.Context, req *backendrpc.GetUserRequest,
) (*backendrpc.GetUserResponse, error) {
	ctx, span := p.tracer.Start(ctx, "PrepareToCallClient")
	span.SetAttributes(attribute.Int64("user.id", req.Id))
	span.AddEvent("some-event", trace.WithAttributes(attribute.String("person.age", "1200")))
	defer span.End()

	client := backendrpc.NewBackendServiceClient(p.conn)
	return client.GetUser(ctx, req)
}
