package rpc

import (
	"context"

	"github.com/LdDl/horizon/rpc/protos_pb"
)

// GetIsochrones Implement GetIsochrones() to match interface
func (ts *Microservice) GetIsochrones(ctx context.Context, in *protos_pb.IsochronesRequest) (*protos_pb.IsochronesResponse, error) {
	return &protos_pb.IsochronesResponse{}, nil
}
