package physics

import "testing"

func TestSweptCircleHitsThinSegment(t *testing.T) {
	wall := LineCollider{ID: "wall", Segment: Segment{A: V(-20, 0), B: V(20, 0)}}
	hit, ok := SweepCircleSegment(V(0, -100), V(0, 200), 5, wall)
	if !ok {
		t.Fatal("fast circle tunnelled through segment")
	}
	if !closeTo(hit.TOI, 0.475) || !hit.Normal.AlmostEqual(V(0, -1)) {
		t.Fatalf("unexpected hit: %+v", hit)
	}
}

func TestSweptCircleHitsSegmentEndCap(t *testing.T) {
	wall := LineCollider{Segment: Segment{A: V(0, 0), B: V(10, 0)}}
	hit, ok := SweepCircleSegment(V(15, -10), V(0, 20), 5, wall)
	if !ok || !closeTo(hit.TOI, 0.5) {
		t.Fatalf("expected tangent end-cap hit, got %+v, %v", hit, ok)
	}
}

func TestSweptCircleCircle(t *testing.T) {
	collider := CircleCollider{Center: V(20, 0), Radius: 4}
	hit, ok := SweepCircleCircle(V(0, 0), V(30, 0), 2, collider)
	if !ok || !closeTo(hit.TOI, 14.0/30.0) || !hit.Normal.AlmostEqual(V(-1, 0)) {
		t.Fatalf("unexpected hit: %+v, %v", hit, ok)
	}
}

func TestWorldIgnoresSeparatingOverlaps(t *testing.T) {
	tests := []struct {
		name     string
		world    World
		ball     Ball
		velocity Vec
		wantPos  Vec
	}{
		{
			name: "circle",
			world: World{Circles: []CircleCollider{{
				ID: "bumper", Center: Vec{}, Radius: 5,
			}}},
			ball:     NewBall(V(6, 0), 2),
			velocity: V(100, 0),
			wantPos:  V(7, 0),
		},
		{
			name: "capsule",
			world: World{Lines: []LineCollider{{
				ID: "wall", Segment: Segment{A: V(-10, 0), B: V(10, 0)}, Radius: 1,
			}}},
			ball:     NewBall(V(0, -2), 2),
			velocity: V(0, -100),
			wantPos:  V(0, -3),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.ball.Velocity = test.velocity
			contacts := test.world.StepBall(&test.ball, 0.01)
			if len(contacts) != 0 {
				t.Fatalf("separating overlap emitted contacts: %+v", contacts)
			}
			if !test.ball.Position.AlmostEqual(test.wantPos) {
				t.Fatalf("separating ball position = %+v, want unobstructed motion to %+v", test.ball.Position, test.wantPos)
			}
		})
	}
}

func TestMaterialRestitutionUsesGeometricMean(t *testing.T) {
	ball := NewBall(Vec{}, 1)
	ball.Restitution = 0.25
	restitution, _ := combineMaterial(&ball, Material{Restitution: 1})
	if !closeTo(restitution, 0.5) {
		t.Fatalf("combined restitution = %v, want geometric mean 0.5", restitution)
	}
}

func TestResolutionRestitutionFrictionAndCorrection(t *testing.T) {
	ball := NewBall(V(0, 0), 2)
	ball.Velocity = V(10, -20)
	ball.Restitution = 0.5
	ball.Friction = 1
	impulse := ResolveStaticContact(&ball, V(0, 1), Vec{}, Material{Restitution: 0.5, Friction: 1})
	if impulse <= 0 || !closeTo(ball.Velocity.Y, 10) || !closeTo(ball.Velocity.X, 0) {
		t.Fatalf("resolved velocity=%+v impulse=%v", ball.Velocity, impulse)
	}
	CorrectPenetration(&ball, V(0, 1), 3)
	if ball.Position.Y <= 3 {
		t.Fatalf("penetration was not corrected: %+v", ball.Position)
	}
}

func TestWorldStepPreventsTunnellingAndBounces(t *testing.T) {
	world := World{Lines: []LineCollider{{
		ID: "floor", Segment: Segment{A: V(-100, 0), B: V(100, 0)},
		Material: Material{Restitution: 1},
	}}}
	ball := NewBall(V(0, -50), 2)
	ball.Restitution = 1
	ball.Velocity = V(0, 1000)
	contacts := world.StepBall(&ball, 0.1)
	if len(contacts) == 0 || contacts[0].ColliderID != "floor" {
		t.Fatalf("expected floor contact, got %+v", contacts)
	}
	if ball.Velocity.Y >= 0 || ball.Position.Y >= -2 {
		t.Fatalf("ball did not bounce above floor: position=%+v velocity=%+v", ball.Position, ball.Velocity)
	}
}
