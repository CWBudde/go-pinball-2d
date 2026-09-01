package physics

import (
	"math"
	"testing"
)

func closeTo(a, b float64) bool { return math.Abs(a-b) < 1e-7 }

func TestVectorAndGeometry(t *testing.T) {
	v := V(3, 4)
	if !closeTo(v.Length(), 5) {
		t.Fatalf("length = %v", v.Length())
	}
	if got := v.Normalized(); !closeTo(got.Length(), 1) {
		t.Fatalf("normalized length = %v", got.Length())
	}
	if got := V(20, 0).ClampLength(7); !got.AlmostEqual(V(7, 0)) {
		t.Fatalf("clamped vector = %+v", got)
	}
	s := Segment{A: V(2, 3), B: V(12, 3)}
	if got := s.ClosestPoint(V(8, 10)); !got.AlmostEqual(V(8, 3)) {
		t.Fatalf("closest point = %+v", got)
	}
	if got := DistancePointSegment(V(-1, 3), s); !closeTo(got, 3) {
		t.Fatalf("endpoint distance = %v", got)
	}
}
