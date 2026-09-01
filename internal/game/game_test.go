package game

import (
	"math"
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
)

const fullChargeDuration = 17.0 / 12.0

type recordingStore struct {
	highScore int
	saves     []int
}

func (s *recordingStore) LoadHighScore() int { return s.highScore }

func (s *recordingStore) SaveHighScore(score int) {
	s.saves = append(s.saves, score)
	s.highScore = score
}

func newStartedGame(t *testing.T) *Game {
	t.Helper()
	g := New(table.New(), &MemoryStore{})
	if g.State != Loading {
		t.Fatalf("initial state = %v, want Loading", g.State)
	}
	g.FinishLoading()
	g.Update(FixedStep, Input{StartPressed: true})
	if g.State != BallReady || g.BallsRemaining != 3 || g.BallNumber != 1 {
		t.Fatalf("start = state %v, remaining %d, ball %d", g.State, g.BallsRemaining, g.BallNumber)
	}
	return g
}

func launchBall(t *testing.T, g *Game) {
	t.Helper()
	advanceFor(g, fullChargeDuration, Input{Plunger: true})
	if g.PlungerCharge < 1 {
		t.Fatalf("plunger charge = %v, want full charge", g.PlungerCharge)
	}
	g.Update(1.0/60.0, Input{})
	if g.State != Playing || g.Ball.Velocity.Y >= 0 {
		t.Fatalf("launch = state %v velocity %+v", g.State, g.Ball.Velocity)
	}
}

func advanceFor(g *Game, seconds float64, input Input) {
	for seconds > 0 {
		dt := math.Min(seconds, 1.0/60.0)
		g.Update(dt, input)
		seconds -= dt
	}
}

func TestNewRejectsDefinitionWithoutTwoFlippers(t *testing.T) {
	definition := table.New()
	definition.Flippers = definition.Flippers[:1]
	defer func() {
		if recover() == nil {
			t.Fatal("New accepted a table definition with fewer than two flippers")
		}
	}()
	New(definition, &MemoryStore{})
}

func TestFlipperControlsFollowNamedDefinitions(t *testing.T) {
	definition := table.New()
	left := definition.Flippers[0]
	right := definition.Flippers[1]
	definition.Flippers = []*physics.Flipper{right, left}

	g := New(definition, &MemoryStore{})
	g.FinishLoading()
	events := g.Update(FixedStep, Input{LeftFlipper: true})
	if !left.Engaged || right.Engaged {
		t.Fatalf("left input engaged flippers: left=%t right=%t", left.Engaged, right.Engaged)
	}
	if !hasEventWithID(events, FlipperFired, "flipper_left") {
		t.Fatalf("left input events = %#v, want named left-flipper event", events)
	}
}

func TestThreeBallStateTransitionsAndBonus(t *testing.T) {
	g := newStartedGame(t)
	for ball := 1; ball <= 3; ball++ {
		launchBall(t, g)
		g.Bonus = 100
		g.BonusMultiplier = 2
		before := g.Score
		g.Ball.Position = physics.V(360, 1060)
		g.Ball.Velocity = physics.Vec{}
		events := g.Update(FixedStep, Input{})
		if g.State != BallLost || g.Score-before != 200 {
			t.Fatalf("ball %d drain = state %v bonus %d", ball, g.State, g.Score-before)
		}
		if !hasEvent(events, BallDrained) || !hasEvent(events, BonusAwarded) {
			t.Fatalf("ball %d missing drain/bonus event: %#v", ball, events)
		}
		advanceFor(g, ballLostDelay+.05, Input{})
		if ball < 3 && g.State != BallReady {
			t.Fatalf("after ball %d = %v, want BallReady", ball, g.State)
		}
	}
	if g.State != GameOver || g.BallsRemaining != 0 {
		t.Fatalf("final state = %v with %d balls", g.State, g.BallsRemaining)
	}
	g.Update(FixedStep, Input{StartPressed: true})
	if g.State != BallReady || g.Score != 0 || g.BallsRemaining != 3 {
		t.Fatalf("restart did not create a new game: %+v", g)
	}
}

func TestTargetBankMultiplierAndLitJackpot(t *testing.T) {
	g := newStartedGame(t)
	launchBall(t, g)
	g.litLanes["rollover_left"] = true
	g.hitTarget(g.Table.DropTargets[0].ID)
	if g.Score != table.DropTargetScore+jackpotPoints {
		t.Fatalf("score = %d, want target plus jackpot", g.Score)
	}
	if len(g.litLanes) != 0 {
		t.Fatal("jackpot did not consume lit lanes")
	}
	for _, target := range g.Table.DropTargets[1:] {
		g.hitTarget(target.ID)
	}
	if g.BonusMultiplier != 2 {
		t.Fatalf("multiplier = %d, want 2", g.BonusMultiplier)
	}
	for _, target := range g.Table.DropTargets {
		if !g.TargetDown(target.ID) {
			t.Fatalf("target %s is not down", target.ID)
		}
	}
}

func TestCompletedTargetBankDoesNotLeakAcrossBallsOrGames(t *testing.T) {
	g := newStartedGame(t)
	g.setState(Playing)
	for _, target := range g.Table.DropTargets {
		g.hitTarget(target.ID)
	}
	if g.bankReset <= 0 {
		t.Fatal("completed target bank did not schedule a reset")
	}

	g.drainBall()
	if g.bankReset != 0 {
		t.Fatalf("bank reset timer after drain = %v, want 0", g.bankReset)
	}
	for _, target := range g.Table.DropTargets {
		if g.TargetDown(target.ID) {
			t.Fatalf("target %s remained down after draining a completed bank", target.ID)
		}
	}
	advanceFor(g, ballLostDelay+.05, Input{})
	if g.State != BallReady || g.bankReset != 0 {
		t.Fatalf("next ball = state %v bank reset %v", g.State, g.bankReset)
	}

	g.bankReset = .5
	g.targetsDown[g.Table.DropTargets[0].ID] = true
	g.rebuildTargetColliders()
	g.startGame()
	if g.bankReset != 0 || len(g.targetsDown) != 0 {
		t.Fatalf("new game retained target bank state: timer %v targets %#v", g.bankReset, g.targetsDown)
	}
}

func TestPauseFreezesSimulation(t *testing.T) {
	g := newStartedGame(t)
	launchBall(t, g)
	g.Update(FixedStep, Input{PausePressed: true})
	position := g.Ball.Position
	advanceFor(g, 1, Input{})
	if g.State != Paused || g.Ball.Position != position {
		t.Fatalf("paused game moved from %+v to %+v", position, g.Ball.Position)
	}
	g.Update(FixedStep, Input{PausePressed: true})
	if g.State != Playing {
		t.Fatalf("resume state = %v", g.State)
	}
}

func TestWeakLaunchCanBeRechargedAndRelaunched(t *testing.T) {
	g := newStartedGame(t)
	g.Update(FixedStep, Input{Plunger: true})
	g.Update(FixedStep, Input{})
	if g.State != Playing {
		t.Fatalf("weak launch state = %v, want Playing", g.State)
	}

	advanceFor(g, 2, Input{})
	if g.State != BallReady {
		t.Fatalf("returned ball state = %v, want BallReady", g.State)
	}
	if g.BallsRemaining != 3 || g.BallNumber != 1 {
		t.Fatalf("returned ball counted as lost: remaining %d, ball %d", g.BallsRemaining, g.BallNumber)
	}
	if !g.Ball.Position.AlmostEqual(g.Table.BallSpawn) || g.Ball.Velocity.LengthSquared() != 0 {
		t.Fatalf("returned ball not seated on plunger: position %+v velocity %+v", g.Ball.Position, g.Ball.Velocity)
	}

	launchBall(t, g)
	exitedShooterLane := false
	for range 60 * 5 {
		g.Update(1.0/60.0, Input{})
		if g.Ball.Position.Y < 300 {
			exitedShooterLane = true
		}
	}
	if !exitedShooterLane {
		t.Fatalf("fully charged relaunch did not exit shooter lane: position %+v", g.Ball.Position)
	}
	if g.Score == 0 {
		t.Fatal("fully charged relaunch completed without scoring")
	}
}

func TestHighScoreStoreFlushesOnceWhenGameEnds(t *testing.T) {
	store := &recordingStore{highScore: 1234}
	g := New(table.New(), store)
	if g.HighScore != 1234 {
		t.Fatalf("loaded high score = %d", g.HighScore)
	}
	g.FinishLoading()
	g.Update(FixedStep, Input{StartPressed: true})
	g.addScore(1300)
	g.addScore(100)
	if g.HighScore != 1400 {
		t.Fatalf("buffered high score = %d, want 1400", g.HighScore)
	}
	if len(store.saves) != 0 {
		t.Fatalf("saved high scores before game end = %v", store.saves)
	}

	g.BallsRemaining = 0
	g.setState(BallLost)
	advanceFor(g, ballLostDelay+.05, Input{})
	if g.State != GameOver {
		t.Fatalf("state after final ball = %v, want GameOver", g.State)
	}
	if len(store.saves) != 1 || store.saves[0] != 1400 {
		t.Fatalf("saved high scores at game end = %v, want [1400]", store.saves)
	}
	advanceFor(g, 1, Input{})
	if len(store.saves) != 1 {
		t.Fatalf("saved high score more than once after game end: %v", store.saves)
	}
}

func TestSeparatingBumperContactDoesNotScore(t *testing.T) {
	g := newStartedGame(t)
	contact := physics.Contact{ColliderID: "bumper_left", Impulse: 0}
	g.scoreContacts([]physics.Contact{contact})
	if g.Score != 0 {
		t.Fatalf("separating zero-impulse bumper contact scored %d points", g.Score)
	}
	if events := g.events.drain(); hasEvent(events, BumperHit) {
		t.Fatalf("separating zero-impulse bumper contact emitted hit event: %#v", events)
	}

	contact.Impulse = 1
	g.scoreContacts([]physics.Contact{contact})
	if g.Score != table.BumperScore {
		t.Fatalf("positive-impulse bumper contact scored %d, want %d", g.Score, table.BumperScore)
	}
}

func TestScriptedSimulationStaysFinite(t *testing.T) {
	g := newStartedGame(t)
	launchBall(t, g)
	exitedShooterLane := false
	for frame := range 60 * 20 {
		input := Input{LeftFlipper: frame%47 < 5, RightFlipper: frame%61 < 7}
		g.Update(1.0/60.0, input)
		if g.Ball.Position.Y < 300 {
			exitedShooterLane = true
		}
		if !g.Ball.Position.IsFinite() || !g.Ball.Velocity.IsFinite() {
			t.Fatalf("non-finite ball at frame %d: %+v", frame, g.Ball)
		}
		if math.Abs(g.Ball.Position.X) > 1000 || math.Abs(g.Ball.Position.Y) > 1400 {
			t.Fatalf("escaped ball at frame %d: %+v", frame, g.Ball.Position)
		}
		if g.State == BallReady {
			launchBall(t, g)
		}
	}
	if !exitedShooterLane {
		t.Fatalf("scripted ball did not exit shooter lane: position %+v", g.Ball.Position)
	}
	if g.Score == 0 {
		t.Fatal("scripted simulation completed without scoring")
	}
}

func TestScriptedSimulationIsConsistentAcrossRefreshRates(t *testing.T) {
	type result struct {
		state             State
		score             int
		position          physics.Vec
		velocity          physics.Vec
		exitedShooterLane bool
	}
	run := func(refreshRate int) result {
		g := newStartedGame(t)
		for range refreshRate * 17 / 12 {
			g.Update(1/float64(refreshRate), Input{Plunger: true})
		}
		g.Update(0, Input{})
		exitedShooterLane := false
		for range refreshRate * 5 {
			g.Update(1/float64(refreshRate), Input{})
			if g.Ball.Position.Y < 300 {
				exitedShooterLane = true
			}
		}
		return result{
			state: g.State, score: g.Score, position: g.Ball.Position,
			velocity: g.Ball.Velocity, exitedShooterLane: exitedShooterLane,
		}
	}

	want := run(60)
	if !want.exitedShooterLane {
		t.Fatalf("60 Hz simulation did not exit shooter lane: %+v", want)
	}
	if want.score == 0 {
		t.Fatalf("60 Hz simulation completed without scoring: %+v", want)
	}
	for _, refreshRate := range []int{120, 144} {
		got := run(refreshRate)
		if !got.exitedShooterLane {
			t.Errorf("%d Hz simulation did not exit shooter lane: %+v", refreshRate, got)
		}
		if got.score == 0 {
			t.Errorf("%d Hz simulation completed without scoring: %+v", refreshRate, got)
		}
		if got.state != want.state || got.score != want.score || !got.position.AlmostEqual(want.position) || !got.velocity.AlmostEqual(want.velocity) {
			t.Errorf("%d Hz result differs from 60 Hz:\n got  %+v\n want %+v", refreshRate, got, want)
		}
	}
}

func hasEvent(events []Event, kind EventKind) bool {
	for _, event := range events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}

func hasEventWithID(events []Event, kind EventKind, id string) bool {
	for _, event := range events {
		if event.Kind == kind && event.ID == id {
			return true
		}
	}
	return false
}
