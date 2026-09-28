package rpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
)

func BenchmarkMapMatchBoundaryGeometry(b *testing.B) {
	for _, observations := range []int{4, 100} {
		points := make([][2]float64, observations)
		for i := range points {
			points[i] = [2]float64{37.66310, 55.77330 + 0.0007*float64(i)/float64(observations-1)}
		}
		b.Run(fmt.Sprintf("RPC/observations_%d", observations), func(b *testing.B) {
			f := newBoundaryGeometryFixture(b, points)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := f.service.RunMapMatch(context.Background(), f.rpcRequest); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("REST/observations_%d", observations), func(b *testing.B) {
			f := newBoundaryGeometryFixture(b, points)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				request := httptest.NewRequest("POST", "/mapmatch", bytes.NewReader(f.restBody))
				request.Header.Set("Content-Type", "application/json")
				response, err := f.app.Test(request)
				if err != nil {
					b.Fatal(err)
				}
				_, err = io.Copy(io.Discard, response.Body)
				response.Body.Close()
				if err != nil || response.StatusCode != 200 {
					b.Fatalf("REST status %d: %v", response.StatusCode, err)
				}
			}
		})
	}
}
