package table

import (
	"fmt"
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
)

func TestNeonRelayFeatureCountsAndIDs(t *testing.T) {
	d := New()
	checks := []struct {
		name string
		got  int
		want int
	}{
		{"bumpers", len(d.Bumpers), 3},
		{"slingshots", len(d.Slingshots), 2},
		{"rollover lanes", len(d.RolloverLanes), 3},
		{"drop targets", len(d.DropTargets), 4},
		{"inlanes", len(d.Inlanes), 2},
		{"outlanes", len(d.Outlanes), 2},
		{"posts", len(d.Posts), 8},
		{"flippers", len(d.Flippers), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s count = %d, want %d", check.name, check.got, check.want)
		}
	}

	ids := make(map[string]string)
	add := func(kind, id string) {
		t.Helper()
		if id == "" {
			t.Errorf("%s has an empty ID", kind)
			return
		}
		if previous, exists := ids[id]; exists {
			t.Errorf("ID %q is shared by %s and %s", id, previous, kind)
		}
		ids[id] = kind
	}
	for _, b := range d.Bumpers {
		add("bumper", b.ID)
	}
	for _, s := range d.Slingshots {
		add("slingshot", s.ID)
	}
	for _, l := range d.RolloverLanes {
		add("rollover", l.ID)
	}
	for _, target := range d.DropTargets {
		add("drop target", target.ID)
	}
	for _, l := range d.Inlanes {
		add("inlane", l.ID)
	}
	for _, l := range d.Outlanes {
		add("outlane", l.ID)
	}
	for _, p := range d.Posts {
		add("post", p.ID)
	}
	for _, f := range d.Flippers {
		add("flipper", f.ID)
	}
	add("drain", d.Drain.ID)
	add("plunger", d.Plunger.ID)
}

func TestDefinitionBuildsCompletePhysicsData(t *testing.T) {
	d := New()
	world := d.World()
	wantLines := len(d.OuterWalls) + len(d.ShooterLane) + len(d.GuideWalls) + len(d.Slingshots) + len(d.DropTargets)
	if len(world.Lines) != wantLines {
		t.Fatalf("world line count = %d, want %d", len(world.Lines), wantLines)
	}
	if got, want := len(world.Circles), len(d.Bumpers)+len(d.Posts); got != want {
		t.Fatalf("world circle count = %d, want %d", got, want)
	}
	wantSensors := len(d.RolloverLanes) + len(d.DropTargets) + len(d.Inlanes) + len(d.Outlanes) + 1
	if got := len(d.Sensors()); got != wantSensors {
		t.Fatalf("sensor count = %d, want %d", got, wantSensors)
	}
	if !world.SafePosition.AlmostEqual(d.BallSpawn) || len(world.Flippers) != 2 {
		t.Fatalf("world did not retain spawn/flippers: %+v", world)
	}
	if impulse := d.Plunger.Mechanism; !impulse.Direction.AlmostEqual(physics.V(0, -1)) {
		t.Fatalf("plunger direction = %+v", impulse.Direction)
	}

	other := New()
	d.Flippers[0].SetEngaged(true)
	if other.Flippers[0].Engaged {
		t.Fatal("New returned shared mutable flipper state")
	}
}

func TestRestingFlippersLeaveBallDrainGap(t *testing.T) {
	d := New()
	leftTip := d.Flippers[0].Tip()
	rightTip := d.Flippers[1].Tip()
	freeGap := rightTip.Sub(leftTip).Length() - d.Flippers[0].Radius - d.Flippers[1].Radius
	if freeGap <= BallRadius*2 {
		t.Fatalf("resting flippers close the drain: free gap %.2f, ball diameter %.2f", freeGap, BallRadius*2)
	}
}

func TestPlayfieldRoutesDroppedBallsThroughDrainSensor(t *testing.T) {
	const (
		step              = 1.0 / 240.0
		maxSimulationTime = 30.0
		emergencyDrainY   = Height + 80
	)
	d := New()
	drain := d.Drain.Sensor()
	rows := []struct {
		y, minX, maxX float64
	}{
		{y: 850, minX: 80, maxX: 560},
		{y: 900, minX: 100, maxX: 550},
		{y: 950, minX: 150, maxX: 500},
	}
	for _, row := range rows {
		for x := row.minX; x <= row.maxX; x += 50 {
			y := row.y
			start := physics.V(x, y)
			t.Run(fmt.Sprintf("x%.0f_y%.0f", x, y), func(t *testing.T) {
				world := d.World()
				ball := physics.NewBall(start, BallRadius)
				for elapsed := 0.0; elapsed < maxSimulationTime; elapsed += step {
					world.StepBall(&ball, step)
					if drain.Overlaps(ball) {
						return
					}
					if ball.Position.Y > emergencyDrainY {
						t.Fatalf("ball bypassed drain sensor and reached emergency fallback at %+v", ball.Position)
					}
				}
				t.Fatalf("ball did not reach drain sensor within %v seconds: position=%+v velocity=%+v", maxSimulationTime, ball.Position, ball.Velocity)
			})
		}
	}
}

func TestTableGeometryStaysInsideLogicalBounds(t *testing.T) {
	d := New()
	point := func(label string, p physics.Vec, margin float64) {
		t.Helper()
		if p.X-margin < 0 || p.X+margin > Width || p.Y-margin < 0 || p.Y+margin > Height {
			t.Errorf("%s at %+v with margin %v is outside %vx%v", label, p, margin, Width, Height)
		}
	}
	segment := func(label string, s physics.Segment, margin float64) {
		t.Helper()
		point(label+" A", s.A, margin)
		point(label+" B", s.B, margin)
	}

	for _, wall := range d.LineColliders() {
		segment(wall.ID, wall.Segment, wall.Radius)
	}
	for _, b := range d.Bumpers {
		point(b.ID, b.Center, b.Radius)
	}
	for _, p := range d.Posts {
		point(p.ID, p.Center, p.Radius)
	}
	for _, lane := range d.RolloverLanes {
		segment(lane.ID, lane.Segment, lane.Radius)
	}
	for _, lane := range d.Inlanes {
		segment(lane.ID, lane.Segment, lane.Radius)
	}
	for _, lane := range d.Outlanes {
		segment(lane.ID, lane.Segment, lane.Radius)
	}
	for _, sling := range d.Slingshots {
		for _, p := range sling.Triangle {
			point(sling.ID, p, 0)
		}
	}
	for _, f := range d.Flippers {
		point(f.ID+" pivot", f.Pivot, f.Radius)
		point(f.ID+" rest tip", f.Tip(), f.Radius)
	}
	point("ball spawn", d.BallSpawn, BallRadius)
	point("plunger", d.Plunger.Position, 0)
	point("drain min", d.Drain.Min, 0)
	point("drain max", d.Drain.Max, 0)
}

func TestColliderAndSensorIDsMatchFeatures(t *testing.T) {
	d := New()
	for _, bumper := range d.Bumpers {
		if bumper.Collider().ID != bumper.ID {
			t.Errorf("bumper collider ID mismatch for %q", bumper.ID)
		}
	}
	for _, sling := range d.Slingshots {
		if sling.Collider().ID != sling.ID {
			t.Errorf("slingshot collider ID mismatch for %q", sling.ID)
		}
	}
	for _, target := range d.DropTargets {
		if target.Collider().ID != target.ID || target.Sensor().ID != target.ID {
			t.Errorf("drop-target physics ID mismatch for %q", target.ID)
		}
	}
	for _, sensor := range d.Sensors() {
		if sensor.SensorID() == "" {
			t.Error("sensor has an empty ID")
		}
	}
}
