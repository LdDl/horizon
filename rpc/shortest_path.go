package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/LdDl/horizon"
	"github.com/LdDl/horizon/rpc/protos_pb"
	"github.com/golang/geo/s2"
)

// GetSP Implement GetSP() to match interface
func (ts *Microservice) GetSP(ctx context.Context, in *protos_pb.SPRequest) (*protos_pb.SPResponse, error) {
	if len(in.Gps) != 2 {
		return nil, fmt.Errorf("please provide 2 GPS points only. Provided: %d", len(in.Gps))
	}

	response := &protos_pb.SPResponse{
		Data:     []*protos_pb.EdgeInfo{},
		Warnings: []string{},
	}

	statesRadiusMeters := 25.0
	if in.StateRadius != nil && *in.StateRadius >= 7 && *in.StateRadius <= 50 {
		statesRadiusMeters = *in.StateRadius
	} else {
		response.Warnings = append(response.Warnings, "stateRadius either nil or not in range [7,50]. Using default value: 25.0")
	}

	gpsMeasurements := horizon.GPSMeasurements{}
	ut := time.Now().UTC().Unix()
	for i := range in.Gps {
		gpsMeasurement := horizon.NewGPSMeasurementFromID(int(ut), in.Gps[i].Lon, in.Gps[i].Lat, 4326)
		gpsMeasurements = append(gpsMeasurements, gpsMeasurement)
		ut++
	}
	result, err := ts.matcher.FindShortestPath(gpsMeasurements[0], gpsMeasurements[1], statesRadiusMeters)
	if err != nil {
		return nil, fmt.Errorf("something went wrong on server side: %v", err)
	}
	for i := range result.Observations {
		observationResult := result.Observations[i]
		feature := &protos_pb.EdgeInfo{
			EdgeId: observationResult.MatchedEdge.ID,
			Weight: observationResult.MatchedEdge.Weight,
		}
		if observationResult.MatchedEdge.Polyline != nil {
			geomLen := len(*observationResult.MatchedEdge.Polyline)
			feature.Geom = make([]*protos_pb.GeoPoint, geomLen)
			for k := range *observationResult.MatchedEdge.Polyline {
				latLng := s2.LatLngFromPoint((*observationResult.MatchedEdge.Polyline)[k])
				feature.Geom[k] = &protos_pb.GeoPoint{
					Lon: latLng.Lng.Degrees(),
					Lat: latLng.Lat.Degrees(),
				}
			}
		}
		response.Data = append(response.Data, feature)

		for j := range observationResult.NextEdges {
			edgeFeature := &protos_pb.EdgeInfo{
				EdgeId: observationResult.NextEdges[j].ID,
				Weight: observationResult.NextEdges[j].Weight,
			}
			if observationResult.NextEdges[j].Geom != nil {
				geomLen := len(observationResult.NextEdges[j].Geom)
				edgeFeature.Geom = make([]*protos_pb.GeoPoint, geomLen)
				for k := range observationResult.NextEdges[j].Geom {
					latLng := s2.LatLngFromPoint(observationResult.NextEdges[j].Geom[k])
					edgeFeature.Geom[k] = &protos_pb.GeoPoint{
						Lon: latLng.Lng.Degrees(),
						Lat: latLng.Lat.Degrees(),
					}
				}
			}
			response.Data = append(response.Data, edgeFeature)
		}
	}
	return response, nil
}
