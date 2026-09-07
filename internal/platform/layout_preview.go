package platform

import (
	"math"

	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
	"github.com/gonutz/prototype/draw"
)

// RenderLayoutFrame draws an untextured blockout from the actual table geometry.
// It is a composition aid, not an alternative gameplay renderer.
func RenderLayoutFrame(window draw.Window, current *game.Game) {
	r := new(renderer)
	width, height := window.Size()
	view := newViewport(width, height)
	window.FillRect(0, 0, width, height, draw.RGB(.07, .08, .09))
	solid := draw.RGB(.4, .43, .45)
	for _, line := range current.Table.LineColliders() {
		r.thickLine(window, view, line.Segment, solid, int(math.Round(line.Radius*2)))
	}
	for _, circle := range current.Table.CircleColliders() {
		x, y := view.point(circle.Center)
		radius := view.size(circle.Radius)
		window.FillEllipse(x-radius, y-radius, radius*2, radius*2, solid)
	}
	// Fill triangle interiors with parallel spans in logical coordinates.
	for _, sling := range current.Table.Slingshots {
		a, b, c := sling.Triangle[0], sling.Triangle[1], sling.Triangle[2]
		for i := 0; i <= 240; i++ {
			t := float64(i) / 240
			r.thickLine(window, view, physics.Segment{A: a.Add(b.Sub(a).Mul(t)), B: a.Add(c.Sub(a).Mul(t))}, solid, 2)
		}
	}
	for _, f := range current.World.Flippers {
		r.thickLine(window, view, physics.Segment{A: f.Pivot, B: f.Tip()}, draw.LightGray, int(f.Radius*2))
	}
	r.centerText(window, "NEON RELAY", view.x(table.TitleOrigin.X+110), view.y(table.TitleOrigin.Y+35), float32(view.scale), draw.LightGray)
	RenderCollisionOverlay(window, current)
}

// RenderCollisionOverlay marks live collision capsules (cyan), sensors (amber),
// flippers (magenta), and the ball (white). Dropped targets have no solid outline.
func RenderCollisionOverlay(window draw.Window, current *game.Game) {
	r := new(renderer)
	width, height := window.Size()
	view := newViewport(width, height)
	circle := func(center physics.Vec, radius float64, col draw.Color) {
		x, y := view.point(center)
		size := view.size(radius)
		r.thickEllipse(window, x-size, y-size, size*2, size*2, col, view.stroke(1))
	}
	capsule := func(segment physics.Segment, radius float64, col draw.Color) {
		direction := segment.B.Sub(segment.A).Normalized()
		normal := physics.V(-direction.Y, direction.X).Mul(radius)
		for _, offset := range []physics.Vec{normal, normal.Mul(-1)} {
			r.thickLine(window, view, physics.Segment{A: segment.A.Add(offset), B: segment.B.Add(offset)}, col, 1)
		}
		circle(segment.A, radius, col)
		circle(segment.B, radius, col)
	}
	for _, line := range current.World.Lines {
		capsule(line.Segment, line.Radius, cyan)
	}
	for _, c := range current.World.Circles {
		circle(c.Center, c.Radius, cyan)
	}
	for _, f := range current.World.Flippers {
		capsule(physics.Segment{A: f.Pivot, B: f.Tip()}, f.Radius, magenta)
	}
	for _, lane := range current.Table.RolloverLanes {
		capsule(lane.Segment, lane.Radius, amber)
	}
	for _, target := range current.Table.DropTargets {
		capsule(target.Segment, target.Sensor().Radius, amber)
	}
	drain := current.Table.Drain
	r.thickRect(window, view.x(drain.Min.X), view.y(drain.Min.Y), view.size(drain.Max.X-drain.Min.X), view.size(drain.Max.Y-drain.Min.Y), amber, view.stroke(1))
	if current.Ball.Active {
		circle(current.Ball.Position, current.Ball.Radius, draw.White)
	}
}
