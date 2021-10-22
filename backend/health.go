package backend

import (
	"context"
	backendrpc "grpcotel/rpc/backend"
)

// HealthServer ...
type HealthServer struct {
	backendrpc.UnsafeHealthServiceServer
}

// Live ...
func (*HealthServer) Live(context.Context, *backendrpc.LiveRequest) (*backendrpc.LiveResponse, error) {
	return &backendrpc.LiveResponse{}, nil
}

// Ready ...
func (*HealthServer) Ready(context.Context, *backendrpc.ReadyRequest) (*backendrpc.ReadyResponse, error) {
	return &backendrpc.ReadyResponse{}, nil
}
