package backend

import (
	"context"
	backendrpc "grpcotel/rpc/backend"
)

// Server ...
type Server struct {
	backendrpc.UnimplementedBackendServiceServer
}

// GetUser ...
func (*Server) GetUser(
	_ context.Context, _ *backendrpc.GetUserRequest,
) (*backendrpc.GetUserResponse, error) {
	return &backendrpc.GetUserResponse{
		Msg: "Some message",
	}, nil
}
