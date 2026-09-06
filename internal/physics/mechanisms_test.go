package physics

import (
	"math"
	"testing"
)

func TestFlipperMotionAndSurfaceImpulse(t *testing.T) {
	f := NewFlipper("left", Vec{}, 10, 1, 0, math.Pi/4)
	f.RiseSpeed = math.Pi
	f.SetEngaged(true)
	f.Step(0.1)
	if f.Angle <= 0 || !closeTo(f.AngularVelocity, math.Pi) {
		t.Fatalf("flipper did not advance: %+v", f)
	}

	point := V(10, 0)
	surface := f.SurfaceVelocity(point)
	if surface.Y <= 0 {
		t.Fatalf("expected upward-moving surface, got %+v", surface)
	}
	ball := NewBall(V(10, 2), 1)
	ball.Restitution = 0.5
	ball.Velocity = V(0, -1)
	ResolveStaticContact(&ball, V(0, 1), surface, f.Material)
	if ball.Velocity.Y <= surface.Y {
		t.Fatalf("moving flipper did not kick ball: ball=%+v surface=%+v", ball.Velocity, surface)
	}
}

func TestFlipperLaunchStaysBelowMaxSpeed(t *testing.T) {
	f := NewFlipper("left", Vec{}, 105, 16, 0, -math.Pi/4)
	f.AngularVelocity = -f.RiseSpeed
	ball := NewBall(V(105, -29), 14)
	world := World{Flippers: []*Flipper{f}}

	contacts := world.StepBall(&ball, 1.0/240)
	if len(contacts) == 0 || contacts[0].Impulse <= 0 {
		t.Fatalf("moving flipper did not launch resting ball: contacts=%+v ball=%+v", contacts, ball)
	}
	if speed := ball.Velocity.Length(); speed <= 0 || speed >= ball.MaxSpeed {
		t.Fatalf("flipper launch speed = %v, want between zero and MaxSpeed %v", speed, ball.MaxSpeed)
	}
}

func TestPlungerChargeLimitsAndRelease(t *testing.T) {
	p := NewPlunger(V(0, -10))
	p.ChargeRate = 2
	p.MinImpulse = 100
	p.MaxImpulse = 500
	p.Hold(10)
	if p.Charge != 1 {
		t.Fatalf("charge should clamp to one, got %v", p.Charge)
	}
	impulse := p.Release()
	if !impulse.AlmostEqual(V(0, -500)) || p.Charge != 0 {
		t.Fatalf("release=%+v remaining charge=%v", impulse, p.Charge)
	}
	if got := p.Release(); got.LengthSquared() != 0 {
		t.Fatalf("empty release=%+v", got)
	}
}

func TestFlipperStepRejectsInvalidInputs(t *testing.T) {
	for _, dt := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		f := NewFlipper("test", Vec{}, 10, 1, 0, 1)
		f.AngularVelocity = 7
		f.Step(dt)
		if f.Angle != 0 || f.AngularVelocity != 0 {
			t.Errorf("Step(%v) = angle %v velocity %v", dt, f.Angle, f.AngularVelocity)
		}
	}
	for _, speed := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		f := NewFlipper("test", Vec{}, 10, 1, 0, 1)
		f.RiseSpeed = speed
		f.SetEngaged(true)
		f.Step(.1)
		if f.Angle != 0 || f.AngularVelocity != 0 {
			t.Errorf("invalid rise speed %v moved flipper: %+v", speed, f)
		}
	}
	for _, angle := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		f := NewFlipper("test", Vec{}, 10, 1, .25, 1)
		f.Angle = angle
		f.Step(.1)
		if !finite(f.Angle) || !finite(f.AngularVelocity) {
			t.Errorf("invalid angle %v was not repaired: %+v", angle, f)
		}
	}
}

func TestPlungerHoldRejectsInvalidInputs(t *testing.T) {
	for _, dt := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		p := NewPlunger(V(0, -1))
		p.Charge = .25
		p.Hold(dt)
		if p.Charge != .25 {
			t.Errorf("Hold(%v) changed charge to %v", dt, p.Charge)
		}
	}
	for _, rate := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		p := NewPlunger(V(0, -1))
		p.Charge = .25
		p.ChargeRate = rate
		p.Hold(.5)
		if p.Charge != .25 {
			t.Errorf("invalid charge rate %v changed charge to %v", rate, p.Charge)
		}
	}
	p := NewPlunger(V(0, -1))
	p.Charge = math.NaN()
	p.Hold(.5)
	if !finite(p.Charge) || p.Charge <= 0 {
		t.Fatalf("non-finite charge was not repaired: %v", p.Charge)
	}
}

func TestSensorsAndCooldown(t *testing.T) {
	ball := NewBall(V(5, 5), 1)
	drain := BoxSensor{ID: "drain", Min: V(0, 4), Max: V(10, 6)}
	target := SegmentSensor{ID: "target", Segment: Segment{A: V(20, 0), B: V(20, 10)}, Radius: 1}
	if !drain.Overlaps(ball) || target.Overlaps(ball) {
		t.Fatalf("sensor overlap mismatch: drain=%v target=%v", drain.Overlaps(ball), target.Overlaps(ball))
	}

	var cooldowns Cooldowns
	if !cooldowns.Allow("lane:ball-1", 0.2) || cooldowns.Allow("lane:ball-1", 0.2) {
		t.Fatal("cooldown did not suppress persistent contact")
	}
	cooldowns.Advance(0.19)
	if cooldowns.Ready("lane:ball-1") {
		t.Fatal("cooldown expired early")
	}
	cooldowns.Advance(0.02)
	if !cooldowns.Ready("lane:ball-1") {
		t.Fatal("cooldown did not expire")
	}
}

func TestBallGuardRepairsNaNAndCapsSpeed(t *testing.T) {
	ball := NewBall(V(math.NaN(), 4), 3)
	ball.Velocity = V(math.Inf(1), 2)
	if !ball.Guard(V(9, 8)) {
		t.Fatal("guard did not report invalid state")
	}
	if !ball.Position.AlmostEqual(V(9, 8)) || ball.Velocity.LengthSquared() != 0 {
		t.Fatalf("guard repair failed: %+v", ball)
	}
	ball.Velocity = V(10000, 0)
	ball.MaxSpeed = 250
	ball.Guard(Vec{})
	if !closeTo(ball.Velocity.Length(), 250) {
		t.Fatalf("velocity not capped: %v", ball.Velocity.Length())
	}
}

func TestBallGuardRepairsDefensiveFields(t *testing.T) {
	for _, radius := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		ball := NewBall(Vec{}, radius)
		ball.Guard(Vec{})
		if ball.Radius != 1 {
			t.Errorf("invalid radius %v repaired to %v, want 1", radius, ball.Radius)
		}
	}
	for _, maxSpeed := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		ball := NewBall(Vec{}, 1)
		ball.MaxSpeed = maxSpeed
		ball.Guard(Vec{})
		if ball.MaxSpeed != DefaultMaxSpeed || !ball.Velocity.IsFinite() {
			t.Errorf("invalid max speed %v repaired to %v", maxSpeed, ball.MaxSpeed)
		}
	}
	ball := NewBall(V(math.NaN(), math.Inf(1)), 1)
	ball.Guard(V(math.Inf(-1), math.NaN()))
	if ball.Position != (Vec{}) {
		t.Fatalf("invalid safe position repaired ball to %+v, want zero", ball.Position)
	}
}
