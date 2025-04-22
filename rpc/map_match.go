package rpc

import (
	"context"

	"github.com/LdDl/horizon/rpc/protos_pb"
)

// RunMapMatch Implement RunMapMatch() to match interface
func (ts *Microservice) RunMapMatch(ctx context.Context, in *protos_pb.MapMatchRequest) (*protos_pb.MapMatchResponse, error) {
	return &protos_pb.MapMatchResponse{}, nil
}
