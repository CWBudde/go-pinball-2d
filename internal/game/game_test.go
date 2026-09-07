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
	assertEvents(t, events, Event{Kind: FlipperFired, ID: "flipper_left", At: left.Pivot})
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
		assertEvents(t, events,
			Event{Kind: BallDrained, At: g.Ball.Position},
			Event{Kind: BonusAwarded, Points: 200},
		)
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

func TestDrainEventPrecedesBonusEvent(t *testing.T) {
	g := newStartedGame(t)
	g.setState(Playing)
	g.Bonus = 100
	g.BonusMultiplier = 2
	g.Ball.Position = physics.V(360, 1060)

	events := g.Update(FixedStep, Input{})
	assertEvents(t, events,
		Event{Kind: BallDrained, At: g.Ball.Position},
		Event{Kind: BonusAwarded, Points: 200},
	)
}

func TestLoadingFallsBackAfterTimeout(t *testing.T) {
	g := New(table.New(), &MemoryStore{})
	advanceFor(g, loadingTimeout-.1, Input{})
	if g.State != Loading {
		t.Fatalf("state before loading timeout = %v, want Loading", g.State)
	}
	advanceFor(g, .2, Input{})
	if g.State != Attract {
		t.Fatalf("state after loading timeout = %v, want Attract", g.State)
	}
}

func TestTargetBankMultiplierAndLitJackpot(t *testing.T) {
	g := newStartedGame(t)
	g.setState(Playing)
	g.World = physics.World{SafePosition: g.Table.BallSpawn}
	g.Ball.Active = true

	lane := g.Table.RolloverLanes[0]
	g.Ball.Position = lane.Segment.A.Add(lane.Segment.B).Mul(.5)
	events := g.Update(FixedStep, Input{})
	assertEvents(t, events, Event{Kind: RolloverLit, ID: lane.ID, Points: lane.Score, At: g.Ball.Position})
	if !g.LaneLit(lane.ID) {
		t.Fatalf("rollover %q was not lit through sensor dispatch", lane.ID)
	}

	target := g.Table.DropTargets[0]
	g.Ball.Position = target.Segment.A.Add(target.Segment.B).Mul(.5)
	events = g.Update(FixedStep, Input{})
	assertEvents(t, events,
		Event{Kind: TargetDown, ID: target.ID, Points: target.Score, At: g.Ball.Position},
		Event{Kind: JackpotAwarded, ID: target.ID, Points: jackpotPoints, At: g.Ball.Position},
	)
	if g.Score != lane.Score+target.Score+jackpotPoints {
		t.Fatalf("score = %d, want rollover, target, and jackpot", g.Score)
	}
	if len(g.litLanes) != 0 {
		t.Fatal("jackpot did not consume lit lanes")
	}
	for i, target := range g.Table.DropTargets[1:] {
		g.Ball.Position = target.Segment.A.Add(target.Segment.B).Mul(.5)
		events = g.Update(FixedStep, Input{})
		want := []Event{{Kind: TargetDown, ID: target.ID, Points: target.Score, At: g.Ball.Position}}
		if i == len(g.Table.DropTargets)-2 {
			want = append(want, Event{Kind: BankCompleted, Points: 2})
		}
		assertEvents(t, events, want...)
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
		if g.Ball.Position.X < 600 && g.Ball.Position.Y < 300 {
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
	assertEvents(t, g.events.drain())

	contact.Impulse = 1
	g.scoreContacts([]physics.Contact{contact})
	if g.Score != table.BumperScore {
		t.Fatalf("positive-impulse bumper contact scored %d, want %d", g.Score, table.BumperScore)
	}
}

func TestContactScoringUsesFeatureMetadataInsteadOfIDPrefixes(t *testing.T) {
	definition := table.New()
	const (
		id    = "main-pop"
		score = 731
	)
	definition.Features[id] = table.Feature{ID: id, Kind: table.FeatureBumper, Score: score}
	g := New(definition, &MemoryStore{})

	point := physics.V(123, 456)
	g.scoreContacts([]physics.Contact{{ColliderID: id, Point: point, Impulse: 1}})
	if g.Score != score {
		t.Fatalf("feature contact score = %d, want %d", g.Score, score)
	}
	events := g.events.drain()
	if len(events) != 1 || events[0].Kind != BumperHit || events[0].ID != id || events[0].Points != score || events[0].At != point {
		t.Fatalf("feature contact events = %#v", events)
	}
}

func TestStepPlayingScoresBumperAndSlingshotContacts(t *testing.T) {
	tests := []struct {
		name      string
		featureID string
		kind      EventKind
		world     physics.World
		score     int
		wantAt    physics.Vec
	}{
		{
			name: "bumper", featureID: "bumper_left", kind: BumperHit, score: table.BumperScore,
			world: physics.World{Circles: []physics.CircleCollider{{
				ID: "bumper_left", Center: physics.Vec{}, Radius: 10,
				Material: physics.Material{Restitution: 1},
			}}},
			wantAt: physics.V(-10, 0),
		},
		{
			name: "slingshot", featureID: "slingshot_left", kind: SlingshotHit, score: table.SlingshotScore,
			world: physics.World{Lines: []physics.LineCollider{{
				ID: "slingshot_left", Segment: physics.Segment{A: physics.V(0, -20), B: physics.V(0, 20)},
				Radius: 10, Material: physics.Material{Restitution: 1},
			}}},
			wantAt: physics.Vec{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := newStartedGame(t)
			g.setState(Playing)
			g.sensors = nil
			g.World = test.world
			g.Ball = physics.NewBall(physics.V(-30, 0), 1)
			g.Ball.Velocity = physics.V(2600, 0)
			g.stepPlaying(.01)
			if g.Score != test.score {
				t.Fatalf("score = %d, want %d", g.Score, test.score)
			}
			events := g.events.drain()
			if len(events) != 1 {
				t.Fatalf("events = %#v, want one contact event", events)
			}
			assertEvents(t, events, Event{Kind: test.kind, ID: test.featureID, Points: test.score, At: test.wantAt})
		})
	}
}

func TestStepPlayingDispatchesDrainAndRecoversOutOfBoundsBall(t *testing.T) {
	t.Run("drain sensor", func(t *testing.T) {
		g := newStartedGame(t)
		g.setState(Playing)
		g.World = physics.World{SafePosition: g.Table.BallSpawn}
		g.Ball.Position = g.Table.Drain.Min.Add(g.Table.Drain.Max).Mul(.5)
		g.stepPlaying(FixedStep)
		if g.State != BallLost || g.BallsRemaining != 2 {
			t.Fatalf("drain = state %v balls %d", g.State, g.BallsRemaining)
		}
		assertEvents(t, g.events.drain(), Event{Kind: BallDrained, At: g.Ball.Position})
	})

	t.Run("out of bounds", func(t *testing.T) {
		g := newStartedGame(t)
		g.setState(Playing)
		g.World = physics.World{SafePosition: g.Table.BallSpawn}
		g.sensors = nil
		g.Ball.Position = physics.V(-101, 500)
		g.Ball.Velocity = physics.V(-20, 3)
		g.stepPlaying(FixedStep)
		if g.Ball.Position != g.Table.BallSpawn || g.Ball.Velocity != (physics.Vec{}) {
			t.Fatalf("out-of-bounds recovery = %+v", g.Ball)
		}
		assertEvents(t, g.events.drain())
	})
}

func TestStepPlayingExpiresTargetBankReset(t *testing.T) {
	g := newStartedGame(t)
	g.setState(Playing)
	g.sensors = nil
	target := g.Table.DropTargets[0]
	g.targetsDown[target.ID] = true
	g.rebuildTargetColliders()
	g.bankReset = FixedStep / 2
	g.Ball.Position = physics.V(360, 600)
	g.Ball.Velocity = physics.Vec{}
	g.stepPlaying(FixedStep)
	if g.bankReset > 0 || len(g.targetsDown) != 0 {
		t.Fatalf("expired bank reset = timer %v targets %#v", g.bankReset, g.targetsDown)
	}
	found := false
	for _, collider := range g.World.Lines {
		found = found || collider.ID == target.ID
	}
	if !found {
		t.Fatalf("target collider %q was not restored", target.ID)
	}
}

func TestScriptedSimulationStaysFinite(t *testing.T) {
	g := newStartedGame(t)
	launchBall(t, g)
	exitedShooterLane := false
	for frame := range 60 * 20 {
		input := Input{LeftFlipper: frame%47 < 5, RightFlipper: frame%61 < 7}
		g.Update(1.0/60.0, input)
		if g.Ball.Position.X < 600 && g.Ball.Position.Y < 300 {
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
			if g.Ball.Position.X < 600 && g.Ball.Position.Y < 300 {
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

func assertEvents(t *testing.T, got []Event, want ...Event) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("events = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("event %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}
