package backend

import (
	"context"
	"google.golang.org/grpc"
	backendrpc "grpcotel/rpc/backend"
)

// Proxy ...
type Proxy struct {
	backendrpc.UnimplementedBackendServiceServer
	conn *grpc.ClientConn
}

// NewProxy ...
func NewProxy(conn *grpc.ClientConn) *Proxy {
	return &Proxy{conn: conn}
}

// GetUser ...
func (p *Proxy) GetUser(
	ctx context.Context, req *backendrpc.GetUserRequest,
) (*backendrpc.GetUserResponse, error) {
	client := backendrpc.NewBackendServiceClient(p.conn)
	return client.GetUser(ctx, req)
}
