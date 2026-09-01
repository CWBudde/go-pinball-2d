package physics

import "math"

// Flipper is a motor-driven rotating capsule. Angle is measured in radians;
// positive angular velocity rotates counter-clockwise in the logical coordinate
// system.
type Flipper struct {
	ID              string
	Pivot           Vec
	Length          float64
	Radius          float64
	Angle           float64
	RestAngle       float64
	ActiveAngle     float64
	RiseSpeed       float64
	ReturnSpeed     float64
	AngularVelocity float64
	Engaged         bool
	Material        Material
}

func NewFlipper(id string, pivot Vec, length, radius, restAngle, activeAngle float64) *Flipper {
	return &Flipper{
		ID: id, Pivot: pivot, Length: length, Radius: radius,
		Angle: restAngle, RestAngle: restAngle, ActiveAngle: activeAngle,
		RiseSpeed: 18, ReturnSpeed: 12,
		Material: Material{Restitution: 0.82, Friction: 0.28},
	}
}

func (f *Flipper) Tip() Vec {
	return f.Pivot.Add(V(math.Cos(f.Angle), math.Sin(f.Angle)).Mul(f.Length))
}

func (f *Flipper) Segment() Segment { return Segment{A: f.Pivot, B: f.Tip()} }

func (f *Flipper) Collider() LineCollider {
	return LineCollider{ID: f.ID, Segment: f.Segment(), Radius: f.Radius, Material: f.Material}
}

func (f *Flipper) SetEngaged(engaged bool) { f.Engaged = engaged }

func (f *Flipper) Step(dt float64) {
	if dt <= 0 || !finite(dt) {
		f.AngularVelocity = 0
		return
	}
	target, speed := f.RestAngle, f.ReturnSpeed
	if f.Engaged {
		target, speed = f.ActiveAngle, f.RiseSpeed
	}
	if speed < 0 || !finite(speed) {
		speed = 0
	}
	old := f.Angle
	difference := target - f.Angle
	step := speed * dt
	if math.Abs(difference) <= step {
		f.Angle = target
	} else {
		f.Angle += math.Copysign(step, difference)
	}
	f.AngularVelocity = (f.Angle - old) / dt
	if !finite(f.AngularVelocity) || !finite(f.Angle) {
		f.Angle = f.RestAngle
		f.AngularVelocity = 0
	}
}

func (f *Flipper) SurfaceVelocity(point Vec) Vec {
	r := point.Sub(f.Pivot)
	return V(-f.AngularVelocity*r.Y, f.AngularVelocity*r.X)
}
