package platform

import (
	"math"

	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
	"github.com/gonutz/prototype/draw"
)

type lightPulse struct {
	age  float64
	at   physics.Vec
	kind game.EventKind
}

type particle struct {
	position, velocity physics.Vec
	age, life          float64
}

// advance owns every animation clock. Drawing never ages or consumes effects.
// Pause freezes mechanisms, flashes, and sparks together with the simulation.
func (r *renderer) advance(current *game.Game, elapsed float64, events []game.Event) {
	dt := math.Max(0, math.Min(elapsed, game.MaxFrame))
	if math.IsNaN(elapsed) || math.IsInf(elapsed, 0) {
		dt = 0
	}
	if current.State == game.Paused {
		dt = 0
	}
	if r.pulses == nil {
		r.pulses = make(map[string]lightPulse)
	}
	if r.targetLift == nil {
		r.targetLift = make(map[string]float64)
	}
	r.elapsedTime += dt
	for id, pulse := range r.pulses {
		pulse.age += dt
		if pulse.age >= .45 {
			delete(r.pulses, id)
		} else {
			r.pulses[id] = pulse
		}
	}
	alive := r.particles[:0]
	for _, spark := range r.particles {
		spark.age += dt
		if spark.age < spark.life {
			alive = append(alive, spark)
		}
	}
	r.particles = alive
	if r.sequenceActive {
		r.sequenceAge += dt
		if r.sequenceAge >= .9 {
			r.sequenceActive = false
		}
	}
	for _, event := range events {
		switch event.Kind {
		case game.GameStarted, game.BallServed:
			clear(r.pulses)
			clear(r.targetLift)
			r.particles = r.particles[:0]
			r.sequenceActive = false
		case game.BumperHit, game.SlingshotHit, game.TargetDown, game.RolloverLit, game.BallLaunched:
			// Feature-keyed pulses coalesce repeated contacts; spark storage is bounded.
			r.pulses[event.ID] = lightPulse{at: event.At, kind: event.Kind}
			if event.Kind == game.BumperHit || event.Kind == game.SlingshotHit || event.Kind == game.TargetDown {
				for i := 0; i < 3 && len(r.particles) < 36; i++ {
					angle := float64(i)*2.399 + event.At.X*.013
					r.particles = append(r.particles, particle{
						position: event.At,
						velocity: physics.V(math.Cos(angle)*85, math.Sin(angle)*85-30), life: .16 + float64(i)*.025,
					})
				}
			}
		case game.BankCompleted, game.JackpotAwarded:
			r.sequenceAge = 0
			r.sequenceActive = true
		}
	}
	for _, target := range current.Table.DropTargets {
		goal := 1.0
		if current.TargetDown(target.ID) {
			goal = 0
		}
		lift, exists := r.targetLift[target.ID]
		if !exists {
			lift = goal
		}
		// The slot stays fixed while the face collapses into it, then rises on reset.
		r.targetLift[target.ID] = lift + physics.Clamp(goal-lift, -dt/.10, dt/.16)
	}
}

func (r *renderer) intensity(id string) float64 {
	pulse, ok := r.pulses[id]
	if !ok {
		return 0
	}
	return math.Pow(math.Max(0, 1-pulse.age/.45), 2)
}

func (r *renderer) sequenceLight(index int) float64 {
	if !r.sequenceActive {
		return 0
	}
	t := r.sequenceAge - float64(index)*.1
	if t < 0 || t > .23 {
		return 0
	}
	return math.Sin(t / .23 * math.Pi)
}

func tint(col draw.Color, alpha float64) draw.Color {
	col.A = float32(physics.Clamp(alpha, 0, 1))
	return col
}

// Small layered pools illuminate the board below the hardware. Their radius
// remains local to the assembly; no expanding rings or whole-table flashes.
func (r *renderer) pool(window draw.Window, view viewport, center physics.Vec, rx, ry, strength float64, col draw.Color) {
	if strength <= .001 {
		return
	}
	x, y := view.point(center)
	for i := 12; i >= 1; i-- {
		scale := float64(i) / 12
		w, h := view.size(rx*2*scale), view.size(ry*2*scale)
		window.FillEllipse(x-w/2, y-h/2, w, h, tint(col, strength*.025*(1-scale)))
	}
}

func (r *renderer) drawLightPools(window draw.Window, current *game.Game, view viewport) {
	for i, b := range current.Table.Bumpers {
		strength := math.Max(r.intensity(b.ID), r.sequenceLight(i+2)*.65)
		r.pool(window, view, b.Center.Add(physics.V(22, 32)), 65, 46, strength, magenta)
		r.pool(window, view, b.Center.Add(physics.V(-36, -20)), 40, 48, strength, cyan)
	}
	for _, s := range current.Table.Slingshots {
		center := s.Triangle[0].Add(s.Triangle[1]).Mul(.5)
		r.pool(window, view, center, 42, 52, r.intensity(s.ID), magenta)
	}
	for i, t := range current.Table.DropTargets {
		r.pool(window, view, t.Segment.A.Add(t.Segment.B).Mul(.5), 43, 27,
			math.Max(r.intensity(t.ID), r.sequenceLight(i)*.8), magenta)
	}
	for _, lane := range current.Table.RolloverLanes {
		r.pool(window, view, lane.Segment.A.Add(lane.Segment.B).Mul(.5), 27, 44, r.intensity(lane.ID), cyan)
	}
}

func (r *renderer) arc(window draw.Window, view viewport, center physics.Vec, rx, ry, start, end float64, col draw.Color, width int) {
	steps := max(2, int(math.Ceil((end-start)*math.Max(rx, ry)/6)))
	for i := 0; i < steps; i++ {
		a, b := start+(end-start)*float64(i)/float64(steps), start+(end-start)*float64(i+1)/float64(steps)
		r.thickLine(window, view, physics.Segment{
			A: center.Add(physics.V(rx*math.Cos(a), ry*math.Sin(a))),
			B: center.Add(physics.V(rx*math.Cos(b), ry*math.Sin(b))),
		}, col, width)
	}
}

func (r *renderer) bumperResponse(window draw.Window, view viewport, b table.Bumper, compression, strength float64) {
	for i, id := range []string{"bumper_left", "bumper_right", "bumper_center"} {
		if b.ID == id {
			strength = math.Max(strength, r.sequenceLight(i+2)*.7)
		}
	}
	if strength <= .001 {
		return
	}
	center := b.Center.Add(physics.V(0, -7+compression))
	// Reflected color on the chrome lip, then a narrow core in the diffuser.
	r.arc(window, view, center, 47, 43, -2.9, -1.7, tint(cyan, strength*.65), 2)
	r.arc(window, view, center, 47, 43, .15, 1.5, tint(magenta, strength*.5), 2)
	for i := 0; i < 20; i++ {
		a := float64(i)*math.Pi/10 + .035
		b := a + math.Pi/10 - .09
		col := magenta
		if i >= 11 && i <= 14 {
			col = cyan
		}
		r.arc(window, view, center, 41.8, 37.8, a, b, tint(col, strength*.22), 5)
		r.arc(window, view, center, 41.8, 37.8, a, b, tint(draw.White, strength*.85), 1)
	}
}

func (r *renderer) laneResponse(window draw.Window, view viewport, lane table.Lane, strength float64) {
	if strength <= .001 {
		return
	}
	p := lane.Segment.A.Add(lane.Segment.B).Mul(.5)
	for _, y := range []float64{-10, 2, 14} {
		for _, sign := range []float64{-1, 1} {
			r.thickLine(window, view, physics.Segment{A: p.Add(physics.V(sign*6, y)), B: p.Add(physics.V(0, y-7))}, tint(draw.White, strength*.85), 1)
		}
	}
}

func (r *renderer) slingResponse(window draw.Window, view viewport, s table.Slingshot) {
	strength := r.intensity(s.ID)
	if strength <= .001 {
		return
	}
	a, b := s.Triangle[0], s.Triangle[1]
	// The kicker glints and recoils inside the existing rubber footprint. The
	// colliding edge and end posts stay fixed, preserving both return clearances.
	mid := a.Add(b).Mul(.5)
	inward := s.Triangle[2].Sub(mid).Normalized()
	offset := inward.Mul(3 * strength)
	edge := physics.Segment{A: a.Add(b.Sub(a).Mul(.13)).Add(offset), B: a.Add(b.Sub(a).Mul(.86)).Add(offset)}
	r.thickLine(window, view, edge, tint(magenta, strength*.4), 7)
	r.thickLine(window, view, edge, tint(draw.White, strength*.9), 1)
}

func (r *renderer) drawTarget(window draw.Window, current *game.Game, view viewport, target table.DropTarget) {
	midpoint := target.Segment.A.Add(target.Segment.B).Mul(.5)
	angle := math.Atan2(target.Segment.B.Y-target.Segment.A.Y, target.Segment.B.X-target.Segment.A.X) - math.Pi/2
	scale := (target.Segment.B.Sub(target.Segment.A).Length() + 2*target.Radius) / table.TargetFrame.ContactLength
	placement := table.TargetFrame.Place(midpoint, scale, angle)
	r.placedSprite(window, "assets/images/target-down.png", view, placement)
	lift, exists := r.targetLift[target.ID]
	if !exists {
		lift = 1
		if current.TargetDown(target.ID) {
			lift = 0
		}
	}
	if lift > .01 {
		// Collapse normal to the long face into its socket without translating or
		// stretching along neighboring targets. The live ball renders over the face.
		placement.Width *= .15 + .85*lift
		r.placedSprite(window, "assets/images/target.png", view, placement)
	}
	strength := r.intensity(target.ID)
	for i, t := range current.Table.DropTargets {
		if t.ID == target.ID {
			strength = math.Max(strength, r.sequenceLight(i))
		}
	}
	if strength > .001 {
		delta := target.Segment.B.Sub(target.Segment.A)
		r.thickLine(window, view, physics.Segment{A: target.Segment.A.Add(delta.Mul(.15)), B: target.Segment.A.Add(delta.Mul(.85))}, tint(magenta, strength*.55), 5)
		r.thickLine(window, view, physics.Segment{A: target.Segment.A.Add(delta.Mul(.25)), B: target.Segment.A.Add(delta.Mul(.75))}, tint(draw.White, strength*.85), 1)
	}
}

func plungerHeadY(charge float64) float64 { return 1003 + 12*physics.Clamp(charge, 0, 1) }

func (r *renderer) drawPlunger(window draw.Window, current *game.Game, view viewport) {
	x := current.Table.Plunger.Position.X
	head := plungerHeadY(current.PlungerCharge)
	// Fixed head size and wire thickness; only coil spacing and exposed rod
	// length change. The spring foot remains inside the well throughout travel.
	r.spriteCentered(window, "assets/images/plunger-rod.png", view, physics.V(x, (head+1032)/2), 6, 1032-head, 0)
	top, bottom := head+5, 1030.0
	for i := 0; i < 6; i++ {
		y := top + (bottom-top)*float64(i)/6
		r.spriteCentered(window, "assets/images/plunger-coil.png", view, physics.V(x, y), 26, 5, 0)
	}
	for _, y := range []float64{head, 1034} {
		r.spriteCentered(window, "assets/images/plunger-head.png", view, physics.V(x, y), 28, 9, 0)
	}
}

func (r *renderer) drawBall(window draw.Window, current *game.Game, view viewport) {
	ball := current.Ball
	r.spriteCentered(window, "assets/images/ball-shadow.png", view, ball.Position.Add(physics.V(3, 5)), 42, 28, 0)
	r.placedSprite(window, "assets/images/ball.png", view, table.BallFrame.Place(ball.Position, ball.Radius/table.BallFrame.ContactRadius, 0))
	// The ball catches a small reflected strip facing nearby active hardware.
	// Iterate feature order rather than a map so compositing is deterministic.
	for _, group := range [][]string{{"bumper_left", "bumper_right", "bumper_center"}, {"target_relay_1", "target_relay_2", "target_relay_3", "target_relay_4"}, {"slingshot_left", "slingshot_right"}} {
		for _, id := range group {
			pulse, ok := r.pulses[id]
			if !ok {
				continue
			}
			delta := pulse.at.Sub(ball.Position)
			strength := r.intensity(id) * math.Max(0, 1-delta.Length()/180)
			if strength <= .01 {
				continue
			}
			angle := math.Atan2(delta.Y, delta.X)
			col := cyan
			if pulse.kind == game.TargetDown || pulse.kind == game.SlingshotHit {
				col = magenta
			}
			r.arc(window, view, ball.Position, ball.Radius-2, ball.Radius-2, angle-.45, angle+.45, tint(col, strength*.9), 1)
		}
	}
}

func (r *renderer) drawEffects(window draw.Window, view viewport) {
	for _, spark := range r.particles {
		p := spark.position.Add(spark.velocity.Mul(spark.age)).Add(physics.V(0, 75*spark.age*spark.age))
		tail := p.Sub(spark.velocity.Normalized().Mul(3))
		strength := 1 - spark.age/spark.life
		r.thickLine(window, view, physics.Segment{A: tail, B: p}, tint(cyan, strength*.85), 1)
	}
}
