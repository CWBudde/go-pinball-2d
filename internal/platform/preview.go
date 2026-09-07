package platform

import (
	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/gonutz/prototype/draw"
)

// PreviewRenderer retains the same effect state as the live app. Advance is
// independent of drawing, allowing sparse captures without slowing simulation.
// The caller passes the events returned by Game.Update exactly once per update.
type PreviewRenderer struct{ renderer renderer }

func (p *PreviewRenderer) Advance(current *game.Game, elapsed float64, events []game.Event) {
	p.renderer.advance(current, elapsed, events)
}

func (p *PreviewRenderer) Draw(window draw.Window, current *game.Game) error {
	r := &p.renderer
	if !r.preload(window) {
		if r.loadError != nil {
			return r.loadError
		}
		return draw.ErrImageLoading
	}
	r.draw(window, current, nil)
	return r.loadError
}

// RenderFrame is a stateless still. Use PreviewRenderer for timed sequences.
func RenderFrame(window draw.Window, current *game.Game) error {
	p := new(PreviewRenderer)
	p.Advance(current, 0, nil)
	return p.Draw(window, current)
}
