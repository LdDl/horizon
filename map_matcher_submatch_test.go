package horizon

import (
	"math"
	"testing"

	"github.com/LdDl/ch"
	"github.com/LdDl/horizon/spatial"
	"github.com/golang/geo/s2"
)

// TestMapMatcherSubMatches tests sub-matching with 4 disconnected road networks
/*
{"type":"FeatureCollection","features":[
{"type":"Feature","properties":{"name":"Network 1: 0=>1","stroke":"#e41a1c","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.57263214574641,62.15710545151077],[90.60075273328482,62.16457203274629]]}},
{"type":"Feature","properties":{"name":"Network 1: 1=>2","stroke":"#e41a1c","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.60075273328482,62.16457203274629],[90.61018533609996,62.151375160382116]]}},
{"type":"Feature","properties":{"name":"Network 1: 2=>3","stroke":"#e41a1c","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.61018533609996,62.151375160382116],[90.62649877200704,62.15977713375611]]}},
{"type":"Feature","properties":{"name":"Network 1: 2=>0","stroke":"#e41a1c","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.61018533609996,62.151375160382116],[90.57263214574641,62.15710545151077]]}},
{"type":"Feature","properties":{"name":"Network 1: 3=>2","stroke":"#e41a1c","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.62649877200704,62.15977713375611],[90.61018533609996,62.151375160382116]]}},
{"type":"Feature","properties":{"name":"Network 1: 3=>1","stroke":"#e41a1c","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.62649877200704,62.15977713375611],[90.60075273328482,62.16457203274629]]}},
{"type":"Feature","properties":{"name":"Network 1: 0=>3","stroke":"#e41a1c","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.57263214574641,62.15710545151077],[90.62649877200704,62.15977713375611]]}},
{"type":"Feature","properties":{"name":"Network 2: 4=>5","stroke":"#377eb8","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.71798405333243,62.162911730393006],[90.74175110552176,62.14217597215935]]}},
{"type":"Feature","properties":{"name":"Network 2: 5=>6","stroke":"#377eb8","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.74175110552176,62.14217597215935],[90.7602917518659,62.170859445110125]]}},
{"type":"Feature","properties":{"name":"Network 2: 6=>7","stroke":"#377eb8","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.7602917518659,62.170859445110125],[90.79169887336894,62.163074983477145]]}},
{"type":"Feature","properties":{"name":"Network 2: 7=>5","stroke":"#377eb8","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.79169887336894,62.163074983477145],[90.74175110552176,62.14217597215935]]}},
{"type":"Feature","properties":{"name":"Network 2: 7=>6","stroke":"#377eb8","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.79169887336894,62.163074983477145],[90.7602917518659,62.170859445110125]]}},
{"type":"Feature","properties":{"name":"Network 2: 6=>4","stroke":"#377eb8","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.7602917518659,62.170859445110125],[90.71798405333243,62.162911730393006]]}},
{"type":"Feature","properties":{"name":"Network 3: 8=>9","stroke":"#4daf4a","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.64672987626358,62.0717130755433],[90.69226990305219,62.095656718913744]]}},
{"type":"Feature","properties":{"name":"Network 3: 9=>10","stroke":"#4daf4a","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.69226990305219,62.095656718913744],[90.71175563865171,62.0777380011518]]}},
{"type":"Feature","properties":{"name":"Network 3: 10=>11","stroke":"#4daf4a","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.71175563865171,62.0777380011518],[90.74158192126038,62.068507519650296]]}},
{"type":"Feature","properties":{"name":"Network 3: 8=>11","stroke":"#4daf4a","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.64672987626358,62.0717130755433],[90.74158192126038,62.068507519650296]]}},
{"type":"Feature","properties":{"name":"Network 3: 11=>9","stroke":"#4daf4a","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.74158192126038,62.068507519650296],[90.69226990305219,62.095656718913744]]}},
{"type":"Feature","properties":{"name":"Network 4: 12=>13","stroke":"#984ea3","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.52273625442604,62.09698255259957],[90.5363378512294,62.08919954674195]]}},
{"type":"Feature","properties":{"name":"Network 4: 13=>14","stroke":"#984ea3","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.5363378512294,62.08919954674195],[90.55839609952318,62.110292596759535]]}},
{"type":"Feature","properties":{"name":"Network 4: 14=>15","stroke":"#984ea3","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.55839609952318,62.110292596759535],[90.58097780458331,62.0985975026332]]}},
{"type":"Feature","properties":{"name":"Network 4: 12=>14","stroke":"#984ea3","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.52273625442604,62.09698255259957],[90.55839609952318,62.110292596759535]]}},
{"type":"Feature","properties":{"name":"Network 4: 13=>15","stroke":"#984ea3","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.5363378512294,62.08919954674195],[90.58097780458331,62.0985975026332]]}},
{"type":"Feature","properties":{"name":"Network 4: 15=>12","stroke":"#984ea3","stroke-width":3},"geometry":{"type":"LineString","coordinates":[[90.58097780458331,62.0985975026332],[90.52273625442604,62.09698255259957]]}},
{"type":"Feature","properties":{"name":"Vertex 0","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.57263214574641,62.15710545151077]}},
{"type":"Feature","properties":{"name":"Vertex 1","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.60075273328482,62.16457203274629]}},
{"type":"Feature","properties":{"name":"Vertex 2","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.61018533609996,62.151375160382116]}},
{"type":"Feature","properties":{"name":"Vertex 3","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.62649877200704,62.15977713375611]}},
{"type":"Feature","properties":{"name":"Vertex 4","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.71798405333243,62.162911730393006]}},
{"type":"Feature","properties":{"name":"Vertex 5","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.74175110552176,62.14217597215935]}},
{"type":"Feature","properties":{"name":"Vertex 6","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.7602917518659,62.170859445110125]}},
{"type":"Feature","properties":{"name":"Vertex 7","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.79169887336894,62.163074983477145]}},
{"type":"Feature","properties":{"name":"Vertex 8","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.64672987626358,62.0717130755433]}},
{"type":"Feature","properties":{"name":"Vertex 9","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.69226990305219,62.095656718913744]}},
{"type":"Feature","properties":{"name":"Vertex 10","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.71175563865171,62.0777380011518]}},
{"type":"Feature","properties":{"name":"Vertex 11","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.74158192126038,62.068507519650296]}},
{"type":"Feature","properties":{"name":"Vertex 12","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.52273625442604,62.09698255259957]}},
{"type":"Feature","properties":{"name":"Vertex 13","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.5363378512294,62.08919954674195]}},
{"type":"Feature","properties":{"name":"Vertex 14","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.55839609952318,62.110292596759535]}},
{"type":"Feature","properties":{"name":"Vertex 15","marker-color":"#808080","marker-size":"small"},"geometry":{"type":"Point","coordinates":[90.58097780458331,62.0985975026332]}},
{"type":"Feature","properties":{"name":"Obs A (near 0=>1)","marker-color":"#ffff00","marker-size":"medium","marker-symbol":"star"},"geometry":{"type":"Point","coordinates":[90.58857670612241,62.161105659415426]}},
{"type":"Feature","properties":{"name":"Obs B (near 2=>3)","marker-color":"#ffff00","marker-size":"medium","marker-symbol":"star"},"geometry":{"type":"Point","coordinates":[90.62401358880209,62.15493544415261]}},
{"type":"Feature","properties":{"name":"Obs C (near 5=>6)","marker-color":"#ffff00","marker-size":"medium","marker-symbol":"star"},"geometry":{"type":"Point","coordinates":[90.74635027813503,62.15888061539417]}},
{"type":"Feature","properties":{"name":"Obs D (near 8=>9)","marker-color":"#ffff00","marker-size":"medium","marker-symbol":"star"},"geometry":{"type":"Point","coordinates":[90.65769436239356,62.08947737076261]}},
{"type":"Feature","properties":{"name":"Obs E (near 9=>10)","marker-color":"#ffff00","marker-size":"medium","marker-symbol":"star"},"geometry":{"type":"Point","coordinates":[90.70219626700805,62.08831745081403]}},
{"type":"Feature","properties":{"name":"Obs F (near 10=>11)","marker-color":"#ffff00","marker-size":"medium","marker-symbol":"star"},"geometry":{"type":"Point","coordinates":[90.72477222344662,62.07547575555694]}},
{"type":"Feature","properties":{"name":"Obs G (near 13=>14)","marker-color":"#ffff00","marker-size":"medium","marker-symbol":"star"},"geometry":{"type":"Point","coordinates":[90.5531896372218,62.1024953047081]}}
]}
*/
func TestMapMatcherSubMatches(t *testing.T) {
	// Observations (GPS measurements)
	gpsMeasurements := GPSMeasurements{
		NewGPSMeasurementFromID(1, 90.58857670612241, 62.161105659415426, 4326), // Near a0-a1 edge
		NewGPSMeasurementFromID(2, 90.62401358880209, 62.15493544415261, 4326),  // Near a2-a3 (a3-a2) edge
		NewGPSMeasurementFromID(3, 90.74635027813503, 62.15888061539417, 4326),  // Near a5-a6 edge
		NewGPSMeasurementFromID(4, 90.65769436239356, 62.08947737076261, 4326),  // Near a8-a9 edge
		NewGPSMeasurementFromID(5, 90.70219626700805, 62.08831745081403, 4326),  // Near a9-a10 edge
		NewGPSMeasurementFromID(6, 90.72477222344662, 62.07547575555694, 4326),  // Near a10-a11 edge
		NewGPSMeasurementFromID(7, 90.5531896372218, 62.1024953047081, 4326),    // Near a13-a14 edge
	}

	// Vertex positions from the graph
	// [lon, lat] format coordinates
	vertices := map[int64][2]float64{
		// Network 1
		0: {90.57263214574641, 62.15710545151077},
		1: {90.60075273328482, 62.16457203274629},
		2: {90.61018533609996, 62.151375160382116},
		3: {90.62649877200704, 62.15977713375611},
		// Network 2
		4: {90.71798405333243, 62.162911730393006},
		5: {90.74175110552176, 62.14217597215935},
		6: {90.7602917518659, 62.170859445110125},
		7: {90.79169887336894, 62.163074983477145},
		// Network 3
		8:  {90.64672987626358, 62.0717130755433},
		9:  {90.69226990305219, 62.095656718913744},
		10: {90.71175563865171, 62.0777380011518},
		11: {90.74158192126038, 62.068507519650296},
		// Network 4
		12: {90.52273625442604, 62.09698255259957},
		13: {90.5363378512294, 62.08919954674195},
		14: {90.55839609952318, 62.110292596759535},
		15: {90.58097780458331, 62.0985975026332},
	}
	// Edge definitions: ID, Source, Target
	type edgeDef struct {
		id     int64
		source int64
		target int64
	}
	edgeDefs := []edgeDef{
		// Network 1 (7 edges)
		{1, 0, 1}, {2, 1, 2}, {3, 2, 3}, {4, 2, 0}, {5, 3, 2}, {6, 3, 1}, {7, 0, 3},
		// Network 2 (6 edges)
		{8, 4, 5}, {9, 5, 6}, {10, 6, 7}, {11, 7, 5}, {12, 7, 6}, {13, 6, 4},
		// Network 3 (5 edges)
		{14, 8, 9}, {15, 9, 10}, {16, 10, 11}, {17, 8, 11}, {18, 11, 9},
		// Network 4 (6 edges)
		{19, 12, 13}, {20, 13, 14}, {21, 14, 15}, {22, 12, 14}, {23, 13, 15}, {24, 15, 12},
	}

	// Populate graph and spatial edges, vertices storage
	graph := ch.Graph{}
	edgesSpatial := []*spatial.Edge{}
	verticesSpatial := []*spatial.Vertex{}
	for vertexID, coords := range vertices {
		err := graph.CreateVertex(vertexID)
		if err != nil {
			t.Errorf("Can't add vertex with id = '%d' to the graph: %v", vertexID, err)
			return
		}
		// Create vertex with s2 point
		s2Point := s2.PointFromLatLng(s2.LatLngFromDegrees(coords[1], coords[0]))
		verticesSpatial = append(verticesSpatial, &spatial.Vertex{
			Point: &s2Point,
			ID:    vertexID,
		})
	}
	for _, edge := range edgeDefs {
		source := vertices[edge.source]
		target := vertices[edge.target]
		sourcePt := s2.LatLngFromDegrees(source[1], source[0])
		targetPt := s2.LatLngFromDegrees(target[1], target[0])

		weight := sourcePt.Distance(targetPt).Radians() * spatial.EarthRadius

		err := graph.AddEdge(edge.source, edge.target, weight)
		if err != nil {
			t.Errorf("Can't add edge from '%d' to '%d' to the graph: %v", edge.source, edge.target, err)
			return
		}
		s2Polyline := s2.Polyline{
			s2.PointFromLatLng(sourcePt),
			s2.PointFromLatLng(targetPt),
		}
		edge := spatial.Edge{
			ID:       edge.id,
			Source:   edge.source,
			Target:   edge.target,
			Weight:   weight,
			Polyline: &s2Polyline,
		}
		edgesSpatial = append(edgesSpatial, &edge)
	}

	// Populate spatial index
	spatialStorage := spatial.NewS2Storage(17, 35)

	// Prepare engine
	mapEngine := NewMapEngine(
		WithGraph(graph),
		WithS2Storage(spatialStorage),
		WithEdges(edgesSpatial),
		WithVertices(verticesSpatial),
	)

	// Create matcher and set engine
	sigma := 10.0
	beta := 2.0
	hmmParams := NewHmmProbabilities(sigma, beta)
	matcher := NewMapMatcher(
		WithHmmParameters(hmmParams),
		WithMapEngine(mapEngine),
	)

	// Define expected results
	// Expected sub-matches:
	// - SubMatch 0 (Network 1): A (0=>1), B (2=>3)
	// - SubMatch 1 (Network 2): C (5=>6)
	// - SubMatch 2 (Network 3): D (8=>9), E (9=>10), F (10=>11)
	// - SubMatch 3 (Network 4): G (13=>14)
	correctStates := MatcherResult{
		SubMatches: []SubMatch{
			{
				Observations: []ObservationResult{
					{Observation: gpsMeasurements[0], MatchedEdge: *mapEngine.edges[0][1]},
					{Observation: gpsMeasurements[1], MatchedEdge: *mapEngine.edges[2][3]},
				},
				Probability: -791.677435,
			},
			{
				Observations: []ObservationResult{
					{Observation: gpsMeasurements[2], MatchedEdge: *mapEngine.edges[5][6]},
				},
				Probability: -954.672372,
			},
			{
				Observations: []ObservationResult{
					{Observation: gpsMeasurements[3], MatchedEdge: *mapEngine.edges[8][9]},
					{Observation: gpsMeasurements[4], MatchedEdge: *mapEngine.edges[9][10]},
					{Observation: gpsMeasurements[5], MatchedEdge: *mapEngine.edges[10][11]},
				},
				Probability: -8100.966270,
			},
			{
				Observations: []ObservationResult{
					{Observation: gpsMeasurements[6], MatchedEdge: *mapEngine.edges[13][14]},
				},
				Probability: -196.691325,
			},
		},
	}

	statesRadiusMeters := 1000.0
	maxStates := 2
	result, err := matcher.Run(gpsMeasurements, statesRadiusMeters, maxStates)
	if err != nil {
		t.Error(err)
		return
	}

	// Check number of sub-matches
	if len(result.SubMatches) != len(correctStates.SubMatches) {
		t.Errorf("Expected %d sub-matches, got %d", len(correctStates.SubMatches), len(result.SubMatches))
		return
	}

	eps := 10e-6
	for s := range result.SubMatches {
		resultSubMatch := result.SubMatches[s]
		correctSubMatch := correctStates.SubMatches[s]

		if len(resultSubMatch.Observations) != len(correctSubMatch.Observations) {
			t.Errorf("SubMatch %d: expected %d observations, got %d",
				s, len(correctSubMatch.Observations), len(resultSubMatch.Observations))
			continue
		}

		if math.Abs(resultSubMatch.Probability-correctSubMatch.Probability) > eps {
			t.Errorf("SubMatch %d: probability should be %f, but got %f",
				s, correctSubMatch.Probability, resultSubMatch.Probability)
		}

		for i := range resultSubMatch.Observations {
			if resultSubMatch.Observations[i].MatchedEdge.Source != correctSubMatch.Observations[i].MatchedEdge.Source ||
				resultSubMatch.Observations[i].MatchedEdge.Target != correctSubMatch.Observations[i].MatchedEdge.Target {
				t.Errorf("SubMatch %d, observation %d: matched edge should be %d=>%d, but got %d=>%d",
					s, resultSubMatch.Observations[i].Observation.id,
					correctSubMatch.Observations[i].MatchedEdge.Source, correctSubMatch.Observations[i].MatchedEdge.Target,
					resultSubMatch.Observations[i].MatchedEdge.Source, resultSubMatch.Observations[i].MatchedEdge.Target,
				)
			}
		}
	}

	// // Generate GeoJSON output for visualization
	// fc := geojson.NewFeatureCollection()

	// // Color palette for sub-matches
	// colors := []string{"#e41a1c", "#377eb8", "#4daf4a", "#984ea3"}
	// for s, subMatch := range result.SubMatches {
	// 	color := colors[s%len(colors)]

	// 	for _, obs := range subMatch.Observations {
	// 		// Add matched edge as LineString
	// 		edgeFeature := S2PolylineToGeoJSONFeature(*obs.MatchedEdge.Polyline)
	// 		edgeFeature.SetProperty("type", "matched_edge")
	// 		edgeFeature.SetProperty("submatch", s)
	// 		edgeFeature.SetProperty("obs_id", obs.Observation.ID())
	// 		edgeFeature.SetProperty("edge", obs.MatchedEdge.Source*1000+obs.MatchedEdge.Target)
	// 		edgeFeature.SetProperty("stroke", color)
	// 		edgeFeature.SetProperty("stroke-width", 8)
	// 		fc.AddFeature(edgeFeature)

	// 		// Add projected point
	// 		projFeature := S2PointToGeoJSONFeature(&obs.ProjectedPoint)
	// 		projFeature.SetProperty("type", "projected_point")
	// 		projFeature.SetProperty("submatch", s)
	// 		projFeature.SetProperty("obs_id", obs.Observation.ID())
	// 		projFeature.SetProperty("marker-color", color)
	// 		projFeature.SetProperty("marker-size", "small")
	// 		projFeature.SetProperty("marker-symbol", "circle")
	// 		fc.AddFeature(projFeature)

	// 		// Add original GPS observation
	// 		obsFeature := obs.Observation.GeoJSON()
	// 		obsFeature.SetProperty("type", "observation")
	// 		obsFeature.SetProperty("submatch", s)
	// 		obsFeature.SetProperty("obs_id", obs.Observation.ID())
	// 		obsFeature.SetProperty("marker-color", "#ffff00")
	// 		obsFeature.SetProperty("marker-size", "medium")
	// 		obsFeature.SetProperty("marker-symbol", "star")
	// 		fc.AddFeature(obsFeature)
	// 	}
	// }

	// // Marshal to JSON string
	// geojsonBytes, err := json.Marshal(fc)
	// if err != nil {
	// 	t.Errorf("Failed to marshal GeoJSON: %v", err)
	// } else {
	// 	t.Logf("Result GeoJSON:\n%s", string(geojsonBytes))
	// }
}
