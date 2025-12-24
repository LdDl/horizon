package rpc

import (
	"context"
	"fmt"

	"github.com/LdDl/horizon"
	"github.com/LdDl/horizon/rpc/protos_pb"
	"github.com/golang/geo/s2"
)

// GetIsochrones Implement GetIsochrones() to match interface
func (ts *Microservice) GetIsochrones(ctx context.Context, in *protos_pb.IsochronesRequest) (*protos_pb.IsochronesResponse, error) {
	response := &protos_pb.IsochronesResponse{
		Isochrones: []*protos_pb.Isochrone{},
		Warnings:   []string{},
	}

	gpsMeasurement := horizon.NewGPSMeasurementFromID(0, in.Lon, in.Lat, 4326)

	maxCost := 0.0
	if in.MaxCost != nil && *in.MaxCost >= 0 {
		maxCost = *in.MaxCost
	} else {
		response.Warnings = append(response.Warnings, "max_cost either nil or not in range [0,+Inf]. Using default value: 0.0")
	}

	maxNearestRadius := horizon.ResolveRadius(in.MaxNearestRadius, horizon.DEFAULT_SP_RADIUS)

	result, err := ts.matcher.FindIsochrones(gpsMeasurement, maxCost, maxNearestRadius)
	if err != nil {
		return nil, err
	}

	for i := range result {
		isochrone := result[i]
		if isochrone.Vertex == nil {
			return nil, fmt.Errorf("empty vertex")
		}
		latLon := s2.LatLngFromPoint(*isochrone.Vertex.Point)
		lon := latLon.Lng.Degrees()
		lat := latLon.Lat.Degrees()
		feature := &protos_pb.Isochrone{
			Id:       int64(i),
			VertexId: isochrone.Vertex.ID,
			Cost:     isochrone.Cost,
			Point: &protos_pb.GeoPoint{
				Lon: lon,
				Lat: lat,
			},
		}
		response.Isochrones = append(response.Isochrones, feature)
	}
	return response, nil
}
