package physics

import "math"

const (
	DefaultSolverIterations = 8
	positionSlop            = 0.001
)

// ResolveStaticContact applies restitution and Coulomb friction against a
// surface that may itself be moving (as a flipper does). The returned value is
// the normal velocity impulse per unit mass.
func ResolveStaticContact(ball *Ball, normal, surfaceVelocity Vec, material Material) float64 {
	n := normal.Normalized()
	if n.LengthSquared() == 0 {
		return 0
	}
	relative := ball.Velocity.Sub(surfaceVelocity)
	vn := relative.Dot(n)
	if vn >= 0 {
		return 0
	}
	restitution, friction := combineMaterial(ball, material)
	normalImpulse := -(1 + restitution) * vn
	ball.Velocity = ball.Velocity.Add(n.Mul(normalImpulse))

	tangent := relative.Sub(n.Mul(vn))
	tangentSpeed := tangent.Length()
	if tangentSpeed > geometryEpsilon {
		frictionImpulse := math.Min(tangentSpeed, friction*normalImpulse)
		ball.Velocity = ball.Velocity.Sub(tangent.Mul(frictionImpulse / tangentSpeed))
	}
	ball.Velocity = ball.Velocity.ClampLength(effectiveMaxSpeed(ball))
	return normalImpulse * ball.Mass
}

func CorrectPenetration(ball *Ball, normal Vec, penetration float64) {
	if penetration > 0 && finite(penetration) {
		ball.Position = ball.Position.Add(normal.Normalized().Mul(penetration + positionSlop))
	}
}

func effectiveMaxSpeed(ball *Ball) float64 {
	if ball.MaxSpeed > 0 && finite(ball.MaxSpeed) {
		return ball.MaxSpeed
	}
	return DefaultMaxSpeed
}

type World struct {
	Gravity       Vec
	Lines         []LineCollider
	Circles       []CircleCollider
	Flippers      []*Flipper
	MaxIterations int
	SafePosition  Vec
}

type collisionCandidate struct {
	hit      Hit
	id       string
	material Material
	velocity Vec
	flipper  bool
}

// StepBall advances one ball by dt. It uses conservative swept collision
// iterations, so a fast ball cannot skip through a thin wall in one step.
func (w *World) StepBall(ball *Ball, dt float64) []Contact {
	if ball == nil || !ball.Active || dt <= 0 || !finite(dt) {
		return nil
	}
	ball.Guard(w.SafePosition)
	ball.Velocity = ball.Velocity.Add(w.Gravity.Mul(dt)).ClampLength(effectiveMaxSpeed(ball))
	remaining := dt
	iterations := w.MaxIterations
	if iterations <= 0 {
		iterations = DefaultSolverIterations
	}
	contacts := make([]Contact, 0, 2)

	for i := 0; i < iterations && remaining > 1e-8; i++ {
		delta := ball.Velocity.Mul(remaining)
		best, found := w.earliestCollision(*ball, delta)
		if !found {
			ball.Position = ball.Position.Add(delta)
			remaining = 0
			break
		}

		toi := Clamp(best.hit.TOI, 0, 1)
		ball.Position = ball.Position.Add(delta.Mul(toi))
		CorrectPenetration(ball, best.hit.Normal, best.hit.Penetration)
		impulse := ResolveStaticContact(ball, best.hit.Normal, best.velocity, best.material)
		contacts = append(contacts, Contact{
			ColliderID: best.id, Point: best.hit.Point, Normal: best.hit.Normal,
			Impulse: impulse, SurfaceVelocity: best.velocity, Flipper: best.flipper,
		})

		// Separate by a tiny amount even for a zero-depth swept hit. This avoids
		// rediscovering the same contact due to roundoff on the next iteration.
		ball.Position = ball.Position.Add(best.hit.Normal.Mul(positionSlop))
		consumed := math.Max(toi, 1e-6)
		remaining *= 1 - consumed
	}
	if remaining > 1e-8 {
		// The iteration cap is an energy/stability guard. Do not blindly advance
		// through unresolved geometry; the next fixed step continues the motion.
		ball.Velocity = ball.Velocity.ClampLength(effectiveMaxSpeed(ball))
	}
	ball.Guard(w.SafePosition)
	return contacts
}

func (w *World) earliestCollision(ball Ball, delta Vec) (collisionCandidate, bool) {
	best := collisionCandidate{hit: Hit{TOI: math.Inf(1)}}
	consider := func(hit Hit, ok bool, id string, material Material, velocity Vec, flipper bool) {
		if ok && hit.TOI < best.hit.TOI {
			best = collisionCandidate{hit: hit, id: id, material: material, velocity: velocity, flipper: flipper}
		}
	}
	for _, line := range w.Lines {
		hit, ok := SweepCircleSegment(ball.Position, delta, ball.Radius, line)
		consider(hit, ok, line.ID, line.Material, Vec{}, false)
	}
	for _, circle := range w.Circles {
		hit, ok := SweepCircleCircle(ball.Position, delta, ball.Radius, circle)
		consider(hit, ok, circle.ID, circle.Material, Vec{}, false)
	}
	for _, flipper := range w.Flippers {
		if flipper == nil {
			continue
		}
		line := flipper.Collider()
		hit, ok := SweepCircleSegment(ball.Position, delta, ball.Radius, line)
		velocity := Vec{}
		if ok {
			velocity = flipper.SurfaceVelocity(hit.Point)
		}
		consider(hit, ok, flipper.ID, flipper.Material, velocity, true)
	}
	return best, !math.IsInf(best.hit.TOI, 1)
}

func (w *World) StepFlippers(dt float64) {
	for _, f := range w.Flippers {
		if f != nil {
			f.Step(dt)
		}
	}
}
