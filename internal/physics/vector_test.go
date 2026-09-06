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

func TestVectorNormalizationAndClampingRejectInvalidValues(t *testing.T) {
	for _, test := range []struct {
		name string
		in   Vec
		want Vec
	}{
		{"zero", Vec{}, Vec{}},
		{"NaN", V(math.NaN(), 1), Vec{}},
		{"positive infinity", V(math.Inf(1), 1), Vec{}},
		{"negative infinity", V(math.Inf(-1), 1), Vec{}},
		{"finite", V(3, 4), V(.6, .8)},
	} {
		t.Run("normalized "+test.name, func(t *testing.T) {
			if got := test.in.Normalized(); !got.AlmostEqual(test.want) {
				t.Fatalf("Normalized() = %+v, want %+v", got, test.want)
			}
		})
	}

	vector := V(3, 4)
	for _, test := range []struct {
		name string
		max  float64
		want Vec
	}{
		{"zero", 0, Vec{}},
		{"negative", -1, Vec{}},
		{"NaN", math.NaN(), Vec{}},
		{"negative infinity", math.Inf(-1), Vec{}},
		{"positive infinity", math.Inf(1), vector},
		{"shorter", 2, V(1.2, 1.6)},
		{"longer", 10, vector},
	} {
		t.Run("clamp "+test.name, func(t *testing.T) {
			if got := vector.ClampLength(test.max); !got.AlmostEqual(test.want) {
				t.Fatalf("ClampLength(%v) = %+v, want %+v", test.max, got, test.want)
			}
		})
	}
	for _, vector := range []Vec{V(math.NaN(), 1), V(math.Inf(1), 1), V(math.Inf(-1), 1)} {
		if got := vector.ClampLength(10); got != (Vec{}) {
			t.Errorf("invalid vector %+v clamped to %+v, want zero", vector, got)
		}
	}
}
