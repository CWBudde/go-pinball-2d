package platform

import (
	"math"
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
)

func TestResponseClocksPauseExpireAndReset(t *testing.T) {
	g := game.New(table.New(), nil)
	g.State = game.Playing
	r := new(renderer)
	event := game.Event{Kind: game.BumperHit, ID: "bumper_left", At: g.Table.Bumpers[0].Center}
	r.advance(g, 0, []game.Event{event, {Kind: game.BankCompleted}})
	if r.intensity(event.ID) != 1 || len(r.particles) != 3 {
		t.Fatal("impact did not start immediately")
	}
	g.State = game.Paused
	for range 30 {
		r.advance(g, .1, nil)
	}
	if r.intensity(event.ID) != 1 || r.particles[0].age != 0 || r.sequenceAge != 0 {
		t.Fatal("pause aged responses")
	}
	g.State = game.Playing
	r.advance(g, .1, nil)
	if r.intensity(event.ID) >= 1 || r.intensity(event.ID) <= 0 {
		t.Fatal("flash did not fade")
	}
	for range 10 {
		r.advance(g, .1, nil)
	}
	if len(r.pulses) != 0 || len(r.particles) != 0 || r.sequenceActive {
		t.Fatal("responses lingered after recovery")
	}
	for range 100 {
		r.advance(g, 0, []game.Event{event})
	}
	if len(r.pulses) != 1 || len(r.particles) > 36 {
		t.Fatal("repeated contacts grew effects without bound")
	}
	r.advance(g, 0, []game.Event{{Kind: game.GameStarted}})
	if len(r.pulses) != 0 || len(r.particles) != 0 {
		t.Fatal("new game retained impact state")
	}
}

func TestResponseClockIsIndependentOfCaptureFrequency(t *testing.T) {
	g := game.New(table.New(), nil)
	g.State = game.Playing
	event := game.Event{Kind: game.BumperHit, ID: "bumper_left"}
	a, b := new(renderer), new(renderer)
	a.advance(g, 0, []game.Event{event})
	b.advance(g, 0, []game.Event{event})
	for range 24 {
		a.advance(g, game.FixedStep, nil)
	}
	b.advance(g, .1, nil)
	if math.Abs(a.intensity(event.ID)-b.intensity(event.ID)) > 1e-9 {
		t.Fatal("response depends on update subdivision")
	}
	before := a.elapsedTime
	for _, dt := range []float64{math.NaN(), math.Inf(1), -1} {
		a.advance(g, dt, nil)
	}
	if a.elapsedTime != before {
		t.Fatal("invalid elapsed time corrupted response clock")
	}
}

func TestTargetDropAndAutomaticResetAnimateActualGameState(t *testing.T) {
	g := game.New(table.New(), nil)
	g.State = game.Playing
	r := new(renderer)
	r.advance(g, 0, nil)
	update := func() { events := g.Update(game.FixedStep, game.Input{}); r.advance(g, game.FixedStep, events) }
	for _, target := range g.Table.DropTargets {
		g.Ball = physics.NewBall(target.Segment.A.Add(target.Segment.B).Mul(.5), table.BallRadius)
		update()
		if !g.TargetDown(target.ID) {
			t.Fatalf("fixture did not drop %s", target.ID)
		}
		if lift := r.targetLift[target.ID]; lift <= 0 || lift >= 1 {
			t.Fatalf("drop skipped its motion: %g", lift)
		}
		for _, line := range g.World.Lines {
			if line.ID == target.ID {
				t.Fatal("visual animation delayed collider removal")
			}
		}
	}
	g.Ball.Position = physics.V(table.PlayfieldCenter, 770)
	g.Ball.Velocity = physics.Vec{}
	g.World.Gravity = physics.Vec{}
	for range 30 {
		update()
	}
	for id, lift := range r.targetLift {
		if lift != 0 {
			t.Fatalf("%s did not settle into socket: %g", id, lift)
		}
	}
	for range 215 {
		update()
	}
	for _, target := range g.Table.DropTargets {
		if g.TargetDown(target.ID) {
			t.Fatal("bank did not reset")
		}
		lift := r.targetLift[target.ID]
		if lift <= 0 || lift >= 1 {
			t.Fatalf("reset skipped its rise: %g", lift)
		}
	}
	for range 45 {
		update()
	}
	for id, lift := range r.targetLift {
		if lift != 1 {
			t.Fatalf("%s did not finish rising: %g", id, lift)
		}
	}
}

func TestPlungerHeadClearsWaitingBallThroughoutCharge(t *testing.T) {
	d := table.New()
	for i := 0; i <= 100; i++ {
		head := plungerHeadY(float64(i) / 100)
		if head-2 <= d.BallSpawn.Y+table.BallRadius || head+5 >= 1030 {
			t.Fatalf("plunger head or spring clips at charge %d", i)
		}
	}
}
