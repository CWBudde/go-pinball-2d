package game

import (
	"strings"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
)

const (
	ballsPerGame  = 3
	jackpotPoints = 5000
	ballLostDelay = 1.25
	gameOverDelay = 6.0
)

// Game contains all mutable simulation and rules state. Visual effects, audio,
// and rendering consume Events and never feed back into this type.
type Game struct {
	State           State
	StateTime       float64
	Score           int
	HighScore       int
	Bonus           int
	BonusMultiplier int
	BallsRemaining  int
	BallNumber      int
	Ball            physics.Ball
	PlungerCharge   float64

	Table *table.Definition
	World physics.World

	clock          Clock
	store          HighScoreStore
	plunger        physics.Plunger
	sensors        []physics.Sensor
	cooldowns      physics.Cooldowns
	targetsDown    map[string]bool
	litLanes       map[string]bool
	resumeState    State
	lastInput      Input
	pendingStart   bool
	pendingPause   bool
	pendingRelease bool
	pendingLeft    bool
	pendingRight   bool
	bankReset      float64
	events         eventQueue
}

func New(def *table.Definition, store HighScoreStore) *Game {
	if def == nil {
		def = table.New()
	}
	if store == nil {
		store = &MemoryStore{}
	}
	g := &Game{
		State: Loading, Table: def, World: def.World(), store: store,
		plunger: def.Plunger.Mechanism, sensors: def.Sensors(),
		targetsDown: make(map[string]bool), litLanes: make(map[string]bool),
		BonusMultiplier: 1,
	}
	g.HighScore = max(0, store.LoadHighScore())
	g.Ball = physics.NewBall(def.BallSpawn, table.BallRadius)
	g.Ball.Active = false
	return g
}

// FinishLoading leaves the loading screen. It is safe to call repeatedly.
func (g *Game) FinishLoading() {
	if g.State == Loading {
		g.setState(Attract)
	}
}

// Update advances with wall-clock elapsed time; the Clock converts it to
// deterministic 1/240 second simulation steps.
func (g *Game) Update(elapsed float64, input Input) []Event {
	g.pendingStart = g.pendingStart || input.StartPressed
	g.pendingPause = g.pendingPause || input.PausePressed
	g.pendingRelease = g.pendingRelease || (g.lastInput.Plunger && !input.Plunger)
	g.pendingLeft = g.pendingLeft || (!g.lastInput.LeftFlipper && input.LeftFlipper)
	g.pendingRight = g.pendingRight || (!g.lastInput.RightFlipper && input.RightFlipper)
	g.lastInput = input

	g.clock.Advance(elapsed, func(dt float64) { g.step(dt) })
	return g.events.drain()
}

func (g *Game) step(dt float64) {
	g.StateTime += dt
	g.cooldowns.Advance(dt)

	if g.pendingStart {
		g.pendingStart = false
		if g.State == Attract || g.State == GameOver {
			g.startGame()
		}
	}
	if g.pendingPause {
		g.pendingPause = false
		switch g.State {
		case Playing, BallReady:
			g.resumeState = g.State
			g.setState(Paused)
		case Paused:
			g.setState(g.resumeState)
		}
	}
	if g.State == Paused || g.State == Loading {
		return
	}

	g.World.Flippers[0].SetEngaged(g.lastInput.LeftFlipper)
	g.World.Flippers[1].SetEngaged(g.lastInput.RightFlipper)
	if g.pendingLeft {
		g.pendingLeft = false
		g.events.emit(Event{Kind: FlipperFired, ID: "flipper_left", At: g.World.Flippers[0].Pivot})
	}
	if g.pendingRight {
		g.pendingRight = false
		g.events.emit(Event{Kind: FlipperFired, ID: "flipper_right", At: g.World.Flippers[1].Pivot})
	}
	g.World.StepFlippers(dt)

	switch g.State {
	case Attract:
		return
	case BallReady:
		g.Ball.Position = g.Table.BallSpawn
		g.Ball.Velocity = physics.Vec{}
		if g.lastInput.Plunger {
			g.plunger.Hold(dt)
		}
		g.PlungerCharge = g.plunger.Charge
		if g.pendingRelease {
			g.pendingRelease = false
			impulse := g.plunger.Release()
			g.PlungerCharge = 0
			if impulse.LengthSquared() > 0 {
				g.Ball.Velocity = g.Ball.Velocity.Add(impulse)
				g.setState(Playing)
				g.events.emit(Event{Kind: BallLaunched, ID: g.Table.Plunger.ID, At: g.Ball.Position})
			}
		}
	case Playing:
		g.stepPlaying(dt)
	case BallLost:
		if g.StateTime >= ballLostDelay {
			if g.BallsRemaining > 0 {
				g.serveBall()
			} else {
				g.setState(GameOver)
				g.events.emit(Event{Kind: GameEnded})
			}
		}
	case GameOver:
		if g.StateTime >= gameOverDelay {
			g.setState(Attract)
		}
	}
}

func (g *Game) stepPlaying(dt float64) {
	if g.bankReset > 0 {
		g.bankReset -= dt
		if g.bankReset <= 0 {
			clear(g.targetsDown)
			g.rebuildTargetColliders()
		}
	}
	contacts := g.World.StepBall(&g.Ball, dt)
	if g.ballReturnedToPlunger() {
		g.rearmPlunger()
		return
	}
	for _, contact := range contacts {
		if !g.cooldowns.Allow("contact:"+contact.ColliderID, .075) {
			continue
		}
		switch {
		case strings.HasPrefix(contact.ColliderID, "bumper_"):
			g.award(table.BumperScore, BumperHit, contact.ColliderID, contact.Point)
		case strings.HasPrefix(contact.ColliderID, "slingshot_"):
			g.award(table.SlingshotScore, SlingshotHit, contact.ColliderID, contact.Point)
		}
	}

	for _, sensor := range g.sensors {
		if !sensor.Overlaps(g.Ball) {
			continue
		}
		id := sensor.SensorID()
		if id == g.Table.Drain.ID {
			g.drainBall()
			return
		}
		switch {
		case strings.HasPrefix(id, "rollover_"):
			if g.cooldowns.Allow("sensor:"+id, .5) {
				g.litLanes[id] = true
				g.award(table.RolloverScore, RolloverLit, id, g.Ball.Position)
			}
		case strings.HasPrefix(id, "target_"):
			g.hitTarget(id)
		}
	}

	if g.Ball.Position.Y > table.Height+80 {
		g.drainBall()
	} else if g.Ball.Position.X < -100 || g.Ball.Position.X > table.Width+100 || g.Ball.Position.Y < -180 {
		// A malformed contact must never poison or permanently lose a game.
		g.Ball.Position = g.Table.BallSpawn
		g.Ball.Velocity = physics.Vec{}
	}
}

func (g *Game) ballReturnedToPlunger() bool {
	spawn := g.Table.BallSpawn
	return g.Ball.Active &&
		g.Ball.Velocity.Y >= 0 &&
		g.Ball.Position.Y >= spawn.Y &&
		g.Ball.Position.X >= spawn.X-30 &&
		g.Ball.Position.X <= spawn.X+30
}

func (g *Game) rearmPlunger() {
	g.Ball.Position = g.Table.BallSpawn
	g.Ball.Velocity = physics.Vec{}
	g.plunger.Charge = 0
	g.PlungerCharge = 0
	g.pendingRelease = false
	g.setState(BallReady)
}

func (g *Game) startGame() {
	g.Score = 0
	g.Bonus = 0
	g.BonusMultiplier = 1
	g.BallsRemaining = ballsPerGame
	g.BallNumber = 0
	clear(g.targetsDown)
	clear(g.litLanes)
	g.cooldowns.Clear()
	g.rebuildTargetColliders()
	g.events.emit(Event{Kind: GameStarted})
	g.serveBall()
}

func (g *Game) serveBall() {
	g.BallNumber++
	g.Ball = physics.NewBall(g.Table.BallSpawn, table.BallRadius)
	g.plunger.Charge = 0
	g.PlungerCharge = 0
	g.pendingRelease = false
	g.setState(BallReady)
	g.events.emit(Event{Kind: BallServed, At: g.Ball.Position})
}

func (g *Game) drainBall() {
	if g.State != Playing {
		return
	}
	g.Ball.Active = false
	g.BallsRemaining--
	points := g.Bonus * g.BonusMultiplier
	if points > 0 {
		g.addScore(points)
		g.events.emit(Event{Kind: BonusAwarded, Points: points})
	}
	g.Bonus = 0
	g.BonusMultiplier = 1
	clear(g.litLanes)
	g.events.emit(Event{Kind: BallDrained, At: g.Ball.Position})
	g.setState(BallLost)
}

func (g *Game) hitTarget(id string) {
	if g.targetsDown[id] || !g.cooldowns.Allow("sensor:"+id, .2) {
		return
	}
	g.targetsDown[id] = true
	g.rebuildTargetColliders()
	g.award(table.DropTargetScore, TargetDown, id, g.Ball.Position)
	if len(g.litLanes) > 0 {
		g.award(jackpotPoints, JackpotAwarded, id, g.Ball.Position)
		clear(g.litLanes)
	}
	if len(g.targetsDown) == len(g.Table.DropTargets) {
		if g.BonusMultiplier < 5 {
			g.BonusMultiplier++
		}
		g.events.emit(Event{Kind: BankCompleted, Points: g.BonusMultiplier})
		g.bankReset = 1
	}
}

func (g *Game) rebuildTargetColliders() {
	lines := g.Table.LineColliders()
	g.World.Lines = g.World.Lines[:0]
	for _, line := range lines {
		if strings.HasPrefix(line.ID, "target_") && g.targetsDown[line.ID] {
			continue
		}
		g.World.Lines = append(g.World.Lines, line)
	}
}

func (g *Game) award(points int, kind EventKind, id string, at physics.Vec) {
	if points <= 0 {
		return
	}
	g.addScore(points)
	g.Bonus += max(10, points/10)
	g.events.emit(Event{Kind: kind, ID: id, Points: points, At: at})
}

func (g *Game) addScore(points int) {
	g.Score += points
	if g.Score > g.HighScore {
		g.HighScore = g.Score
		g.store.SaveHighScore(g.HighScore)
	}
}

func (g *Game) setState(state State) {
	g.State = state
	g.StateTime = 0
}

func (g *Game) TargetDown(id string) bool { return g.targetsDown[id] }
func (g *Game) LaneLit(id string) bool    { return g.litLanes[id] }
