package physics

import "math"

type Segment struct {
	A Vec
	B Vec
}

func (s Segment) Vector() Vec { return s.B.Sub(s.A) }

func (s Segment) ClosestPoint(p Vec) Vec {
	ab := s.Vector()
	denom := ab.LengthSquared()
	if denom <= geometryEpsilon {
		return s.A
	}
	t := Clamp(p.Sub(s.A).Dot(ab)/denom, 0, 1)
	return s.A.Add(ab.Mul(t))
}

func DistancePointSegment(p Vec, s Segment) float64 {
	return p.Sub(s.ClosestPoint(p)).Length()
}

// RayCircle returns the earliest fraction along delta at which the point at
// start enters the circle. Fractions are in [0,1].
func RayCircle(start, delta, center Vec, radius float64) (float64, Vec, bool) {
	radius = math.Max(0, radius)
	m := start.Sub(center)
	c := m.LengthSquared() - radius*radius
	if c <= 0 {
		n := m.Normalized()
		if n.LengthSquared() == 0 {
			n = delta.Mul(-1).Normalized()
			if n.LengthSquared() == 0 {
				n = V(0, -1)
			}
		}
		return 0, n, true
	}
	a := delta.LengthSquared()
	if a <= geometryEpsilon {
		return 0, Vec{}, false
	}
	b := m.Dot(delta)
	disc := b*b - a*c
	if disc < 0 {
		return 0, Vec{}, false
	}
	t := (-b - math.Sqrt(math.Max(0, disc))) / a
	if t < 0 || t > 1 {
		return 0, Vec{}, false
	}
	n := start.Add(delta.Mul(t)).Sub(center).Normalized()
	return t, n, true
}
