package horizon

import (
	"math"
	"testing"

	"github.com/golang/geo/s2"
)

func TestMapMatcherNearestCandidatesOnLargeGraph(t *testing.T) {
	matcher, err := NewMapMatcherFromFiles(NewHmmProbabilities(50, 30), "./test_data/osm2ch_export.csv")
	if err != nil {
		t.Fatal(err)
	}
	query := s2.PointFromLatLng(s2.LatLngFromDegrees(55.750521916863391, 37.600694677555865))
	got, err := matcher.engine.storage.FindNearest(query, 5)
	if err != nil {
		t.Fatal(err)
	}
	// A full segment scan ranks edge 20524 fifth; the old ring search returned edge 42395 at 49.36 m.
	want := []uint64{23141, 38260, 42808, 8302, 20524}
	if len(got) != len(want) {
		t.Fatalf("got %v, want five candidates", got)
	}
	for i, id := range want {
		if got[i].EdgeID != id {
			t.Errorf("candidate %d: got %+v, want edge %d", i, got[i], id)
		}
	}
	if math.Abs(got[4].DistanceTo-41.7309896171835) > 1e-8 {
		t.Errorf("fifth distance: got %.12f m, want 41.730989617184 m", got[4].DistanceTo)
	}
}
