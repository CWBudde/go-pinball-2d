package platform

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/CWBudde/go-pinball-2d/internal/display"
	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/table"
	"github.com/gonutz/prototype/draw"
)

const instrumentFont = "assets/images/instrument-font.png"

// Each glyph uses the same padded atlas on GLFW, WASM and the offline surface.
// Position both edges independently to avoid accumulated rounding at half size.
func (r *renderer) instrumentText(w draw.Window, view viewport, text string, x, y, height float64, tone int) {
	advance := height * display.CellWidth / display.CellHeight
	for i, ch := range []rune(strings.ToUpper(text)) {
		if ch == ' ' {
			continue
		}
		sx, sy := display.Cell(ch, tone)
		left, right := view.x(x+float64(i)*advance), view.x(x+float64(i+1)*advance)
		err := w.DrawImageFilePart(instrumentFont, sx, sy, display.CellWidth, display.CellHeight,
			left, view.y(y), max(1, right-left), view.size(height), 0)
		if err != nil && !errors.Is(err, draw.ErrImageLoading) && r.loadError == nil {
			r.loadError = fmt.Errorf("draw instrument glyph: %w", err)
		}
	}
}

func (r *renderer) instrumentCentered(w draw.Window, view viewport, text string, x, y, height float64, tone int) {
	r.instrumentText(w, view, text, x-display.Width(text, height)/2, y, height, tone)
}

// Scores retain every digit. Long totals reduce type size inside their own
// instrument instead of overlapping HIGH or silently wrapping the display.
func (r *renderer) instrumentNumber(w draw.Window, view viewport, value int, x, y, width, height float64, tone int) {
	text := fmt.Sprintf("%08d", max(0, value))
	height = math.Min(height, width/float64(len(text))*display.CellHeight/display.CellWidth)
	r.instrumentText(w, view, text, x, y, height, tone)
}

func (r *renderer) drawHUD(w draw.Window, current *game.Game, view viewport) {
	r.instrumentText(w, view, "SCORE", 54, 11, 14, display.Silver)
	r.instrumentNumber(w, view, current.Score, 54, 23, 181, 28, display.Cyan)
	r.instrumentText(w, view, "HIGH", 255, 12, 14, display.Silver)
	r.instrumentNumber(w, view, current.HighScore, 255, 28, 106, 19, display.Silver)
	r.instrumentText(w, view, "BALL", 451, 11, 14, display.Silver)
	r.instrumentText(w, view, fmt.Sprintf("%d/3", min(current.BallNumber, 3)), 451, 23, 28, display.Cyan)
	r.instrumentText(w, view, fmt.Sprintf("BONUS X%d", current.BonusMultiplier), 547, 11, 14, display.Amber)
	bonus := fmt.Sprint(current.Bonus)
	h := math.Min(22, 116/float64(len(bonus))*1.5)
	r.instrumentText(w, view, bonus, 547, 26, h, display.Amber)
}

// These are flush playfield displays: render before mechanisms and the ball,
// so even a ball rolling across the apron/status glass remains fully visible.
func (r *renderer) drawTableDisplays(w draw.Window, current *game.Game, view viewport) {
	down := 0
	for _, target := range current.Table.DropTargets {
		if current.TargetDown(target.ID) {
			down++
		}
	}
	label := fmt.Sprintf("%d/4  X%d", down, current.BonusMultiplier)
	if down == len(current.Table.DropTargets) {
		label = "COMPLETE"
	}
	r.instrumentCentered(w, view, label, 511, 677, 17, display.Pink)
	title, hint, tone := stateDisplay(current)
	r.instrumentCentered(w, view, title, table.PlayfieldCenter, 977, 22, tone)
	r.instrumentCentered(w, view, hint, table.PlayfieldCenter, 1002, 16, display.Silver)
	if current.State == game.BallReady {
		// Ten discrete LEDs accompany the moving spring, without a floating bar.
		for i := range 10 {
			col := draw.RGB(.10, .18, .20)
			if float64(i) < current.PlungerCharge*10 {
				col = amber
			}
			w.FillRect(view.x(247+float64(i)*17), view.y(1026), view.size(12), view.size(4), col)
		}
	} else {
		r.instrumentCentered(w, view, "A/D FLIP   P PAUSE", table.PlayfieldCenter, 1023, 14, display.Silver)
	}
}

func stateDisplay(g *game.Game) (title, hint string, tone int) {
	switch g.State {
	case game.Loading:
		return "INITIALIZING", "LOADING RELAY", display.Cyan
	case game.Attract:
		return "NEON RELAY", "ENTER TO CONNECT", display.Cyan
	case game.BallReady:
		if g.PlungerCharge > 0 {
			return fmt.Sprintf("CHARGE %03d%%", int(math.Round(g.PlungerCharge*100))), "RELEASE TO LAUNCH", display.Amber
		}
		return "LAUNCH READY", "HOLD SPACE / DOWN", display.Amber
	case game.Paused:
		return "PAUSED", "P TO RESUME", display.Amber
	case game.BallLost:
		if g.BallsRemaining == 0 {
			return "BALL LOST", "FINAL SCORE ABOVE", display.Pink
		}
		return "BALL LOST", fmt.Sprintf("NEXT BALL %d/3", g.BallNumber+1), display.Pink
	case game.GameOver:
		return "GAME OVER", "ENTER TO RESTART", display.Pink
	case game.Playing:
		for _, lane := range g.Table.RolloverLanes {
			if g.LaneLit(lane.ID) {
				return "RELAY ARMED", "HIT A TARGET +5000", display.Cyan
			}
		}
		return "LINK THE RELAY", "LIGHT LANE + TARGET", display.Cyan
	}
	return "", "", display.Silver
}
