package game

import (
	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
)

const (
	ballsPerGame   = 3
	jackpotPoints  = 5000
	ballLostDelay  = 1.25
	gameOverDelay  = 6.0
	loadingTimeout = 10.0
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
	savedHighScore int
	plunger        physics.Plunger
	leftFlipper    *physics.Flipper
	rightFlipper   *physics.Flipper
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
	leftFlipper, rightFlipper := gameFlippers(def.Flippers)
	if store == nil {
		store = &MemoryStore{}
	}
	g := &Game{
		State: Loading, Table: def, World: def.World(), store: store,
		plunger: def.Plunger.Mechanism, sensors: def.Sensors(),
		leftFlipper: leftFlipper, rightFlipper: rightFlipper,
		targetsDown: make(map[string]bool), litLanes: make(map[string]bool),
		BonusMultiplier: 1,
	}
	g.savedHighScore = max(0, store.LoadHighScore())
	g.HighScore = g.savedHighScore
	g.Ball = physics.NewBall(def.BallSpawn, table.BallRadius)
	g.Ball.Active = false
	return g
}

func gameFlippers(flippers []*physics.Flipper) (*physics.Flipper, *physics.Flipper) {
	if len(flippers) < 2 {
		panic("game: table definition must contain at least two flippers")
	}
	var left, right *physics.Flipper
	for _, flipper := range flippers {
		if flipper == nil {
			continue
		}
		switch flipper.ID {
		case "flipper_left":
			left = flipper
		case "flipper_right":
			right = flipper
		}
	}
	pickOther := func(excluded *physics.Flipper) *physics.Flipper {
		for _, flipper := range flippers {
			if flipper != nil && flipper != excluded {
				return flipper
			}
		}
		return nil
	}
	if left == nil {
		left = pickOther(right)
	}
	if right == nil {
		right = pickOther(left)
	}
	if left == nil || right == nil || left == right {
		panic("game: table definition must contain distinct left and right flippers")
	}
	return left, right
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
	if g.State == Loading {
		if g.StateTime >= loadingTimeout {
			g.setState(Attract)
		}
		return
	}
	if g.State == Paused {
		return
	}

	g.leftFlipper.SetEngaged(g.lastInput.LeftFlipper)
	g.rightFlipper.SetEngaged(g.lastInput.RightFlipper)
	if g.pendingLeft {
		g.pendingLeft = false
		g.events.emit(Event{Kind: FlipperFired, ID: g.leftFlipper.ID, At: g.leftFlipper.Pivot})
	}
	if g.pendingRight {
		g.pendingRight = false
		g.events.emit(Event{Kind: FlipperFired, ID: g.rightFlipper.ID, At: g.rightFlipper.Pivot})
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
				g.flushHighScore()
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
	g.scoreContacts(contacts)

	for _, sensor := range g.sensors {
		if !sensor.Overlaps(g.Ball) {
			continue
		}
		id := sensor.SensorID()
		feature, ok := g.Table.Feature(id)
		if !ok {
			continue
		}
		switch feature.Kind {
		case table.FeatureDrain:
			g.drainBall()
			return
		case table.FeatureRollover:
			if g.cooldowns.Allow("sensor:"+id, .5) {
				g.litLanes[id] = true
				g.award(feature.Score, RolloverLit, id, g.Ball.Position)
			}
		case table.FeatureDropTarget:
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

func (g *Game) scoreContacts(contacts []physics.Contact) {
	for _, contact := range contacts {
		if contact.Impulse <= 0 {
			continue
		}
		feature, ok := g.Table.Feature(contact.ColliderID)
		if !ok || (feature.Kind != table.FeatureBumper && feature.Kind != table.FeatureSlingshot) {
			continue
		}
		if !g.cooldowns.Allow("contact:"+contact.ColliderID, .075) {
			continue
		}
		switch feature.Kind {
		case table.FeatureBumper:
			g.award(feature.Score, BumperHit, contact.ColliderID, contact.Point)
		case table.FeatureSlingshot:
			g.award(feature.Score, SlingshotHit, contact.ColliderID, contact.Point)
		}
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
	g.bankReset = 0
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
	g.events.emit(Event{Kind: BallDrained, At: g.Ball.Position})
	points := g.Bonus * g.BonusMultiplier
	if points > 0 {
		g.addScore(points)
		g.events.emit(Event{Kind: BonusAwarded, Points: points})
	}
	g.Bonus = 0
	g.BonusMultiplier = 1
	if g.bankReset > 0 {
		clear(g.targetsDown)
		g.rebuildTargetColliders()
	}
	g.bankReset = 0
	clear(g.litLanes)
	g.setState(BallLost)
}

func (g *Game) hitTarget(id string) {
	feature, ok := g.Table.Feature(id)
	if !ok || feature.Kind != table.FeatureDropTarget {
		return
	}
	if g.targetsDown[id] || !g.cooldowns.Allow("sensor:"+id, .2) {
		return
	}
	g.targetsDown[id] = true
	g.rebuildTargetColliders()
	g.award(feature.Score, TargetDown, id, g.Ball.Position)
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
	active := lines[:0]
	for _, line := range lines {
		feature, ok := g.Table.Feature(line.ID)
		if ok && feature.Kind == table.FeatureDropTarget && g.targetsDown[line.ID] {
			continue
		}
		active = append(active, line)
	}
	g.World.SetLines(active)
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
	}
}

func (g *Game) flushHighScore() {
	if g.HighScore <= g.savedHighScore {
		return
	}
	g.store.SaveHighScore(g.HighScore)
	g.savedHighScore = g.HighScore
}

func (g *Game) setState(state State) {
	g.State = state
	g.StateTime = 0
}

func (g *Game) TargetDown(id string) bool { return g.targetsDown[id] }
func (g *Game) LaneLit(id string) bool    { return g.litLanes[id] }
