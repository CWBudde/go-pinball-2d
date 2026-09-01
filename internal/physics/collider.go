package physics

import "math"

// Material values are combined with a ball's material during resolution.
type Material struct {
	Restitution float64
	Friction    float64
}

var DefaultMaterial = Material{Restitution: DefaultRestitution, Friction: DefaultFriction}

type LineCollider struct {
	ID       string
	Segment  Segment
	Radius   float64
	Material Material
}

type CircleCollider struct {
	ID       string
	Center   Vec
	Radius   float64
	Material Material
}

type Hit struct {
	TOI         float64
	Normal      Vec
	Point       Vec
	Penetration float64
}

type Contact struct {
	ColliderID      string
	Point           Vec
	Normal          Vec
	Impulse         float64
	SurfaceVelocity Vec
	Flipper         bool
}

func SweepCircleCircle(start, delta Vec, movingRadius float64, collider CircleCollider) (Hit, bool) {
	r := math.Max(0, movingRadius) + math.Max(0, collider.Radius)
	t, n, ok := RayCircle(start, delta, collider.Center, r)
	if !ok {
		return Hit{}, false
	}
	pos := start.Add(delta.Mul(t))
	penetration := math.Max(0, r-pos.Sub(collider.Center).Length())
	return Hit{TOI: t, Normal: n, Point: pos.Sub(n.Mul(movingRadius)), Penetration: penetration}, true
}

// SweepCircleSegment sweeps a circle against a capsule (a segment with an
// optional radius), preventing tunnelling through thin walls.
func SweepCircleSegment(start, delta Vec, movingRadius float64, collider LineCollider) (Hit, bool) {
	r := math.Max(0, movingRadius) + math.Max(0, collider.Radius)
	seg := collider.Segment
	ab := seg.Vector()
	length := ab.Length()
	if length <= geometryEpsilon {
		return SweepCircleCircle(start, delta, movingRadius, CircleCollider{Center: seg.A, Radius: collider.Radius})
	}

	closest := seg.ClosestPoint(start)
	offset := start.Sub(closest)
	if offset.LengthSquared() <= r*r {
		n := offset.Normalized()
		if n.LengthSquared() == 0 {
			n = ab.Perp().Normalized()
			if delta.Dot(n) > 0 {
				n = n.Mul(-1)
			}
		}
		return Hit{TOI: 0, Normal: n, Point: closest, Penetration: math.Max(0, r-offset.Length())}, true
	}

	bestT := math.Inf(1)
	bestNormal := Vec{}
	bestPoint := Vec{}
	dir := ab.Div(length)
	normal := dir.Perp()
	d0 := start.Sub(seg.A).Dot(normal)
	dv := delta.Dot(normal)
	if math.Abs(dv) > geometryEpsilon {
		for _, side := range []float64{-1, 1} {
			t := (side*r - d0) / dv
			if t < 0 || t > 1 || t >= bestT {
				continue
			}
			p := start.Add(delta.Mul(t))
			along := p.Sub(seg.A).Dot(dir)
			if along >= 0 && along <= length && delta.Dot(normal.Mul(side)) < 0 {
				bestT, bestNormal = t, normal.Mul(side)
				bestPoint = seg.A.Add(dir.Mul(along))
			}
		}
	}

	for _, endpoint := range []Vec{seg.A, seg.B} {
		t, n, ok := RayCircle(start, delta, endpoint, r)
		if ok && t < bestT && delta.Dot(n) <= geometryEpsilon {
			bestT, bestNormal, bestPoint = t, n, endpoint.Add(n.Mul(collider.Radius))
		}
	}
	if math.IsInf(bestT, 1) {
		return Hit{}, false
	}
	return Hit{TOI: bestT, Normal: bestNormal, Point: bestPoint}, true
}

func combineMaterial(ball *Ball, surface Material) (float64, float64) {
	restitution := Clamp(math.Max(ball.Restitution, surface.Restitution), 0, 1.2)
	friction := Clamp(math.Sqrt(math.Max(0, ball.Friction)*math.Max(0, surface.Friction)), 0, 1)
	return restitution, friction
}
