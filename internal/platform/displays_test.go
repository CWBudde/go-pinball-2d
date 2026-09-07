package platform

import (
	"image"
	"math"
	"strings"
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/display"
	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/table"
	"github.com/gonutz/prototype/draw"
)

type instrumentRecorder struct {
	draw.Window
	destinations []image.Rectangle
}

func (w *instrumentRecorder) DrawImageFilePart(_ string, _, _, _, _, x, y, width, height, _ int) error {
	w.destinations = append(w.destinations, image.Rect(x, y, x+width, y+height))
	return nil
}

func TestLargeScoresStayInsideInstrument(t *testing.T) {
	for _, size := range []image.Point{{360, 540}, {720, 1080}, {1440, 2160}} {
		v := newViewport(size.X, size.Y)
		w := new(instrumentRecorder)
		r := new(renderer)
		r.instrumentNumber(w, v, math.MaxInt, 54, 23, 181, 28, display.Cyan)
		for _, rect := range w.destinations {
			if !rect.In(image.Rect(v.x(54), v.y(23), v.x(235), v.y(51))) {
				t.Fatalf("large score escaped instrument at %v: %v", size, rect)
			}
		}
		if len(w.destinations) < 10 {
			t.Fatal("large score was truncated")
		}
	}
}

func TestStateInstructionsFitApronAndHaveAtlasGlyphs(t *testing.T) {
	g := game.New(table.New(), nil)
	for _, state := range []game.State{game.Loading, game.Attract, game.BallReady, game.Playing, game.Paused, game.BallLost, game.GameOver} {
		g.State = state
		for _, charge := range []float64{0, .5, 1} {
			g.PlungerCharge = charge
			for _, remaining := range []int{0, 1, 2} {
				g.BallsRemaining = remaining
				title, hint, _ := stateDisplay(g)
				if title == "" || hint == "" || display.Width(title, 22) > 215 || display.Width(hint, 16) > 215 {
					t.Fatalf("%s instructions do not fit: %q / %q", state, title, hint)
				}
				for _, ch := range title + hint {
					if !strings.ContainsRune(display.Characters, ch) {
						t.Fatalf("missing glyph %q", ch)
					}
				}
			}
		}
	}
}
