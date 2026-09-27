package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/LdDl/horizon"
	"github.com/LdDl/horizon/rest"
	"github.com/LdDl/horizon/rpc/protos_pb"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/geo/s2"
)

type boundaryGeometryCase struct {
	name   string
	points [][2]float64
	edges  []int64
}

type boundaryGeometryFixture struct {
	matcher      *horizon.MapMatcher
	service      *Microservice
	app          *fiber.App
	rpcRequest   *protos_pb.MapMatchRequest
	restBody     []byte
	measurements horizon.GPSMeasurements
}

func newBoundaryGeometryFixture(t testing.TB, points [][2]float64) boundaryGeometryFixture {
	t.Helper()
	matcher, err := horizon.NewMapMatcherFromFiles(horizon.NewHmmProbabilities(50, 2), "../test_data/matcher_4326_test.csv")
	if err != nil {
		t.Fatal(err)
	}
	maxStates, rpcMaxStates, radius := 1, int32(1), 7.0
	f := boundaryGeometryFixture{
		matcher: matcher, service: &Microservice{matcher: matcher}, app: fiber.New(),
		rpcRequest: &protos_pb.MapMatchRequest{MaxStates: &rpcMaxStates, StateRadius: &radius},
	}
	f.app.Post("/mapmatch", rest.MapMatch(matcher))
	restRequest := rest.MapMatchRequest{MaxStates: &maxStates, StateRadius: &radius}
	start := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	for i, point := range points {
		timestamp := start.Add(time.Duration(i) * time.Second)
		f.measurements = append(f.measurements, horizon.NewGPSMeasurement(i, point[0], point[1], 4326, horizon.WithGPSTime(timestamp)))
		f.rpcRequest.Gps = append(f.rpcRequest.Gps, &protos_pb.GPSToMapMatch{Tm: timestamp.Format(timestampLayout), Lon: point[0], Lat: point[1]})
		restRequest.Data = append(restRequest.Data, rest.GPSToMapMatch{Timestamp: timestamp.Format(timestampLayout), LonLat: point})
	}
	f.restBody, err = json.Marshal(restRequest)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func geometryCoordinates(points []*protos_pb.GeoPoint) [][]float64 {
	coordinates := make([][]float64, len(points))
	for i, point := range points {
		coordinates[i] = []float64{point.Lon, point.Lat}
	}
	return coordinates
}

func TestMapMatchBoundaryGeometry(t *testing.T) {
	for _, test := range []boundaryGeometryCase{
		{"leading", [][2]float64{{37.66312, 55.77330}, {37.66310, 55.77355}, {37.663095, 55.7740}, {37.66309, 55.77483}}, []int64{1, 1, 1, 3}},
		{"trailing", [][2]float64{{37.66310, 55.77355}, {37.663060, 55.77455}, {37.663075, 55.77475}, {37.663130, 55.7750}}, []int64{1, 3, 3, 3}},
		{"one_edge", [][2]float64{{37.66312, 55.77330}, {37.66310, 55.77355}, {37.663095, 55.7740}}, []int64{1, 1, 1}},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newBoundaryGeometryFixture(t, test.points)
			raw, err := f.matcher.Run(f.measurements, 7, 1)
			if err != nil || len(raw.SubMatches) != 1 || len(raw.SubMatches[0].Observations) != len(test.edges) {
				t.Fatalf("unexpected matcher result: %v, error %v", raw, err)
			}
			var previous rest.MapMatchResponse
			for repeat := 0; repeat < 3; repeat++ {
				rpcResponse, err := f.service.RunMapMatch(context.Background(), f.rpcRequest)
				if err != nil {
					t.Fatal(err)
				}
				request := httptest.NewRequest("POST", "/mapmatch", bytes.NewReader(f.restBody))
				request.Header.Set("Content-Type", "application/json")
				httpResponse, err := f.app.Test(request)
				if err != nil {
					t.Fatal(err)
				}
				var restResponse rest.MapMatchResponse
				err = json.NewDecoder(httpResponse.Body).Decode(&restResponse)
				httpResponse.Body.Close()
				if err != nil || httpResponse.StatusCode != 200 {
					t.Fatalf("REST status %d: %v", httpResponse.StatusCode, err)
				}
				if len(rpcResponse.SubMatches) != 1 || len(restResponse.SubMatches) != 1 {
					t.Fatal("response segmentation changed")
				}
				rpcMatch, restMatch := rpcResponse.SubMatches[0], restResponse.SubMatches[0]
				if rpcMatch.Probability != raw.SubMatches[0].Probability || restMatch.Probability != rpcMatch.Probability {
					t.Fatal("response construction changed the matching score")
				}
				if len(rpcMatch.Observations) != len(test.edges) || len(restMatch.Observations) != len(test.edges) {
					t.Fatal("response observations changed")
				}
				last := len(test.edges) - 1
				for i, edgeID := range test.edges {
					r, h, original := rpcMatch.Observations[i], restMatch.Observations[i], raw.SubMatches[0].Observations[i]
					if !r.IsMatched || !h.IsMatched || r.EdgeId != edgeID || h.EdgeID != edgeID || original.MatchedEdge.ID != edgeID {
						t.Fatalf("observation %d: unexpected matched edge", i)
					}
					point := s2.LatLngFromPoint(original.ProjectedPoint)
					if r.ProjectedPoint.Lon != point.Lng.Degrees() || r.ProjectedPoint.Lat != point.Lat.Degrees() {
						t.Fatal("response construction changed the projection")
					}
					if !reflect.DeepEqual(geometryCoordinates(r.MatchedEdge), h.MatchedEdge.Geometry.LineString) ||
						!reflect.DeepEqual([]float64{r.ProjectedPoint.Lon, r.ProjectedPoint.Lat}, h.ProjectedPoint.Geometry.Point) {
						t.Fatal("REST and RPC geometries differ")
					}
					if edgeID == test.edges[0] && !reflect.DeepEqual(r.MatchedEdge[0], rpcMatch.Observations[0].ProjectedPoint) {
						t.Errorf("observation %d redraws the prefix before the first projection", i)
					}
					if edgeID == test.edges[last] && !reflect.DeepEqual(r.MatchedEdge[len(r.MatchedEdge)-1], rpcMatch.Observations[last].ProjectedPoint) {
						t.Errorf("observation %d redraws the suffix after the last projection", i)
					}
					if (len(r.MatchedEdgeCut) > 0) != (i == 0 || i == last) || (h.MatchedEdgeCut != nil) != (len(r.MatchedEdgeCut) > 0) {
						t.Fatal("removed geometry must occur only at the boundaries")
					}
					if h.MatchedEdgeCut != nil && !reflect.DeepEqual(geometryCoordinates(r.MatchedEdgeCut), h.MatchedEdgeCut.Geometry.LineString) {
						t.Fatal("REST and RPC cuts differ")
					}
					if len(r.NextEdges) != len(h.NextEdges) || len(r.NextEdges) != len(original.NextEdges) {
						t.Fatal("route edge count changed")
					}
					for j, next := range r.NextEdges {
						if next.Id != original.NextEdges[j].ID || next.Weight != original.NextEdges[j].Weight ||
							next.Id != h.NextEdges[j].ID || next.Weight != h.NextEdges[j].Weight ||
							!reflect.DeepEqual(geometryCoordinates(next.Geom), h.NextEdges[j].Geom.Geometry.LineString) {
							t.Fatal("route metadata changed or API geometries differ")
						}
						if next.Id == test.edges[last] && !reflect.DeepEqual(next.Geom[len(next.Geom)-1], rpcMatch.Observations[last].ProjectedPoint) {
							t.Errorf("observation %d next edge %d redraws the terminal suffix", i, j)
						}
					}
				}
				if repeat > 0 && !reflect.DeepEqual(previous, restResponse) {
					t.Fatal("repeated requests changed the response")
				}
				previous = restResponse
			}
		})
	}
}
