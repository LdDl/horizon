package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/LdDl/horizon"
	"github.com/LdDl/horizon/rpc/protos_pb"
	"github.com/golang/geo/s2"
)

var timestampLayout = "2006-01-02T15:04:05"

// RunMapMatch Implement RunMapMatch() to match interface
func (ts *Microservice) RunMapMatch(ctx context.Context, in *protos_pb.MapMatchRequest) (*protos_pb.MapMatchResponse, error) {
	if len(in.Gps) < 3 {
		return nil, fmt.Errorf("please provide 3 GPS points atleast. Provided: %d", len(in.Gps))
	}
	response := &protos_pb.MapMatchResponse{
		Data:     []*protos_pb.ObservationEdge{},
		Warnings: []string{},
	}

	gpsMeasurements := horizon.GPSMeasurements{}
	for i := range in.Gps {
		tm, err := time.Parse(timestampLayout, in.Gps[i].Tm)
		if err != nil {
			return nil, fmt.Errorf("wrong timestamp layout. Please use YYYY-MM-DDTHH:mm:SS")
		}
		// Use index of measurement as ID
		gpsMeasurement := horizon.NewGPSMeasurement(i, in.Gps[i].Lon, in.Gps[i].Lat, 4326, horizon.WithGPSTime(tm))
		gpsMeasurements = append(gpsMeasurements, gpsMeasurement)
	}

	maxStates := 5
	if in.MaxStates != nil && *in.MaxStates > 0 && *in.MaxStates < 10 {
		maxStates = int(*in.MaxStates)
	} else {
		response.Warnings = append(response.Warnings, "maxStates either nil or not in range [1,10]. Using default value: 5")
	}

	statesRadiusMeters := 25.0
	if in.StateRadius != nil && *in.StateRadius >= 7 && *in.StateRadius <= 50 {
		statesRadiusMeters = *in.StateRadius
	} else {
		response.Warnings = append(response.Warnings, "stateRadius either nil or not in range [7,50]. Using default value: 25.0")
	}

	result, err := ts.matcher.Run(gpsMeasurements, statesRadiusMeters, maxStates)
	if err != nil {
		return nil, fmt.Errorf("something went wrong on server side: %v", err)
	}

	response.Data = make([]*protos_pb.ObservationEdge, len(result.Observations))
	for i := range result.Observations {
		observationResult := result.Observations[i]
		if observationResult.MatchedEdge.Polyline == nil {
			return nil, fmt.Errorf("matched edge has nil polyline nil for observation %d", observationResult.Observation.ID())
		}
		matchedEdgePolyline := *observationResult.MatchedEdge.Polyline

		var matchedEdgeCut s2.Polyline
		if i == 0 {
			matchedEdgePolyline, matchedEdgeCut = horizon.ExtractCutUpTo(matchedEdgePolyline, observationResult.ProjectedPoint, observationResult.ProjectionPointIdx)
		} else if i == len(result.Observations)-1 {
			matchedEdgePolyline, matchedEdgeCut = horizon.ExtractCutUpFrom(matchedEdgePolyline, observationResult.ProjectedPoint, observationResult.ProjectionPointIdx)
		}

		if observationResult.MatchedVertex.Point == nil {
			return nil, fmt.Errorf("matched vertex has nil point for observation %d", observationResult.Observation.ID())
		}
		vertexPoint := s2.LatLngFromPoint(*observationResult.MatchedVertex.Point)
		projectedPoint := s2.LatLngFromPoint(observationResult.ProjectedPoint)

		geomLen := len(matchedEdgePolyline)
		line := make([]*protos_pb.GeoPoint, geomLen)
		for k := range matchedEdgePolyline {
			latLng := s2.LatLngFromPoint(matchedEdgePolyline[k])
			line[k] = &protos_pb.GeoPoint{
				Lon: latLng.Lng.Degrees(),
				Lat: latLng.Lat.Degrees(),
			}
		}
		response.Data[i] = &protos_pb.ObservationEdge{
			ObsIdx:      int32(observationResult.Observation.ID()),
			EdgeId:      observationResult.MatchedEdge.ID,
			MatchedEdge: line,
			MatchedVertex: &protos_pb.GeoPoint{
				Lon: vertexPoint.Lng.Degrees(),
				Lat: vertexPoint.Lat.Degrees(),
			},
			ProjectedPoint: &protos_pb.GeoPoint{
				Lon: projectedPoint.Lng.Degrees(),
				Lat: projectedPoint.Lat.Degrees(),
			},
			NextEdges: make([]*protos_pb.IntermediateEdge, len(observationResult.NextEdges)),
		}
		if len(matchedEdgeCut) > 0 {
			cutLine := make([]*protos_pb.GeoPoint, len(matchedEdgeCut))
			for k := range matchedEdgeCut {
				latLng := s2.LatLngFromPoint(matchedEdgeCut[k])
				cutLine[k] = &protos_pb.GeoPoint{
					Lon: latLng.Lng.Degrees(),
					Lat: latLng.Lat.Degrees(),
				}
			}
			response.Data[i].MatchedEdgeCut = cutLine
		}
		for j := range observationResult.NextEdges {
			nextLine := make([]*protos_pb.GeoPoint, len(observationResult.NextEdges[j].Geom))
			for k := range observationResult.NextEdges[j].Geom {
				latLng := s2.LatLngFromPoint(observationResult.NextEdges[j].Geom[k])
				nextLine[k] = &protos_pb.GeoPoint{
					Lon: latLng.Lng.Degrees(),
					Lat: latLng.Lat.Degrees(),
				}
			}
			response.Data[i].NextEdges[j] = &protos_pb.IntermediateEdge{
				Geom:   nextLine,
				Weight: observationResult.NextEdges[j].Weight,
				Id:     observationResult.NextEdges[j].ID,
			}
		}
	}
	return response, nil
}
