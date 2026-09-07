package platform

import (
	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/gonutz/prototype/draw"
)

// RenderFrame draws the production game renderer to an alternate drawing surface.
// It does not start a window, read input, play sound, or access browser globals.
// The caller owns simulation advancement and the destination surface.
func RenderFrame(window draw.Window, current *game.Game) error {
	r := new(renderer)
	if !r.preload(window) {
		if r.loadError != nil {
			return r.loadError
		}
		return draw.ErrImageLoading
	}
	r.draw(window, current, 0, nil)
	return r.loadError
}
