package physics

const (
	DefaultRestitution = 0.72
	DefaultFriction    = 0.12
	DefaultMaxSpeed    = 2600.0
)

// Ball is the only dynamic rigid body needed by the table simulation. Balls
// have unit mass; impulses therefore change velocity directly.
type Ball struct {
	Position    Vec
	Velocity    Vec
	Radius      float64
	Restitution float64
	Friction    float64
	MaxSpeed    float64
	Active      bool
}

func NewBall(position Vec, radius float64) Ball {
	return Ball{
		Position: position, Radius: radius,
		Restitution: DefaultRestitution, Friction: DefaultFriction,
		MaxSpeed: DefaultMaxSpeed, Active: true,
	}
}

// Guard repairs non-finite state and clamps runaway velocity. It returns true
// if invalid position or velocity data was found.
func (b *Ball) Guard(safePosition Vec) bool {
	bad := !b.Position.IsFinite() || !b.Velocity.IsFinite()
	if !b.Position.IsFinite() {
		b.Position = safePosition
	}
	if !b.Velocity.IsFinite() {
		b.Velocity = Vec{}
	}
	if b.Radius <= 0 || !finite(b.Radius) {
		b.Radius = 1
		bad = true
	}
	max := b.MaxSpeed
	if max <= 0 || !finite(max) {
		max = DefaultMaxSpeed
		b.MaxSpeed = max
	}
	b.Velocity = b.Velocity.ClampLength(max)
	return bad
}
