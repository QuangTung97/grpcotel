package backend

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"grpcotel/pkg/level"
	backendrpc "grpcotel/rpc/backend"
	"time"
)

// Server ...
type Server struct {
	backendrpc.UnimplementedBackendServiceServer
	tracer trace.Tracer
}

// NewServer ...
func NewServer(tracer trace.Tracer) *Server {
	return &Server{tracer: tracer}
}

func (s *Server) doSleeping(ctx context.Context) {
	ctx, span := s.tracer.Start(ctx, "Sleeping")

	sc := span.SpanContext()
	fmt.Println("INNER SPAN", sc.TraceID(), sc.SpanID())
	time.Sleep(2 * time.Millisecond)

	level.Extract(ctx).Info("Sleeping")
	level.Extract(ctx).Info("Inside Span")

	span.End()
}

// GetUser ...
func (s *Server) GetUser(
	ctx context.Context, _ *backendrpc.GetUserRequest,
) (*backendrpc.GetUserResponse, error) {
	s.doSleeping(ctx)

	level.Extract(ctx).Info("Outside Span")

	return &backendrpc.GetUserResponse{
		Msg: "Some message",
	}, nil
}
