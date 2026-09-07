package platform

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
	"github.com/gonutz/prototype/draw"
)

var requiredImages = []string{
	"assets/images/background.png",
	"assets/images/playfield-markings.png",
	"assets/images/table-shadows.png",
	"assets/images/table-hardware.png",
	"assets/images/table-foreground.png",
	"assets/images/logo.png",
	"assets/images/favicon.png",
	"assets/images/ball.png",
	"assets/images/flipper.png",
	"assets/images/bumper.png",
	"assets/images/bumper-material.png",
	"assets/images/bumper-patch.png",
	"assets/images/bumper-shadow.png",
	"assets/images/bumper-emission.png",
	"assets/images/post.png",
	"assets/images/target.png",
	"assets/images/target-down.png",
	"assets/images/lane-light.png",
	"assets/images/lane-light-off.png",
	"assets/images/plunger.png",
	"assets/images/glow.png",
	"assets/images/particle.png",
}

const tableOutlineWidth = 3

var (
	ink     = draw.RGB(.02, .03, .09)
	cyan    = draw.RGB(.13, .91, 1)
	magenta = draw.RGB(1, .17, .68)
	lime    = draw.RGB(.66, 1, .31)
	amber   = draw.RGB(1, .75, .22)
	red     = draw.RGB(1, .22, .3)
)

type particle struct {
	position physics.Vec
	life     float64
	maxLife  float64
}

type renderer struct {
	loaded      map[string]bool
	loadError   error
	particles   []particle
	shake       float64
	elapsedTime float64
}

func (r *renderer) preload(window draw.Window) bool {
	if r.loadError != nil {
		return false
	}
	if r.loaded == nil {
		r.loaded = make(map[string]bool, len(requiredImages))
	}
	ready := true
	for _, path := range requiredImages {
		if r.loaded[path] {
			continue
		}
		_, _, err := window.ImageSize(path)
		switch {
		case err == nil:
			r.loaded[path] = true
		case errors.Is(err, draw.ErrImageLoading):
			ready = false
		default:
			r.loadError = fmt.Errorf("load %s: %w", path, err)
			return false
		}
	}
	return ready
}

func (r *renderer) consume(events []game.Event) {
	for _, event := range events {
		switch event.Kind {
		case game.BumperHit, game.SlingshotHit, game.TargetDown, game.JackpotAwarded:
			count := 5
			if event.Kind == game.JackpotAwarded {
				count = 18
				r.shake = .35
			}
			for range count {
				r.particles = append(r.particles, particle{position: event.At, life: .55, maxLife: .55})
			}
		case game.BallDrained:
			r.shake = .22
		}
	}
}

func (r *renderer) draw(window draw.Window, current *game.Game, elapsed float64, statusError error) {
	r.elapsedTime += max(0, elapsed)
	width, height := window.Size()
	window.FillRect(0, 0, width, height, ink)
	view := newViewport(width, height)
	if r.shake > 0 {
		r.shake -= elapsed
		magnitude := 4 * view.scale * math.Min(1, r.shake*5)
		view.offsetX += int(math.Sin(r.elapsedTime*89) * magnitude)
		view.offsetY += int(math.Cos(r.elapsedTime*73) * magnitude)
	}

	r.image(window, "assets/images/background.png", view.offsetX, view.offsetY, view.width, view.height, 0)
	// Flat material patches sit below the printed identity and every mechanism.
	for _, bumper := range current.Table.Bumpers {
		placement := table.BumperMaterialFrame.Place(bumper.Center, bumper.Radius/table.BumperMaterialFrame.ContactRadius, 0)
		r.placedSprite(window, "assets/images/bumper-patch.png", view, placement)
	}
	r.image(window, "assets/images/playfield-markings.png", view.offsetX, view.offsetY, view.width, view.height, 0)
	for _, bumper := range current.Table.Bumpers {
		placement := table.BumperMaterialFrame.Place(bumper.Center, bumper.Radius/table.BumperMaterialFrame.ContactRadius, 0)
		r.placedSprite(window, "assets/images/bumper-shadow.png", view, placement)
	}
	// Layer order: flat playfield, cast shadows, static mechanisms, dynamic
	// mechanisms/ball, safe foreground covers, emission/effects, instrument HUD.
	r.image(window, "assets/images/table-shadows.png", view.offsetX, view.offsetY, view.width, view.height, 0)
	r.image(window, "assets/images/table-hardware.png", view.offsetX, view.offsetY, view.width, view.height, 0)
	r.drawTable(window, current, view)
	r.image(window, "assets/images/table-foreground.png", view.offsetX, view.offsetY, view.width, view.height, 0)
	r.drawEffects(window, view, elapsed)
	r.drawHUD(window, current, view)
	r.drawState(window, current, view)

	visibleError := r.loadError
	if visibleError == nil {
		visibleError = statusError
	}
	if visibleError != nil {
		window.FillRect(view.offsetX, view.offsetY+view.height-view.size(52), view.width, view.size(52), draw.RGBA(.18, .01, .04, .94))
		message := truncate("ERROR: "+visibleError.Error(), 88)
		r.text(window, message, view.offsetX+view.size(14), view.offsetY+view.height-view.size(38), float32(math.Max(.65, view.scale*.8)), red)
	}
}

func (r *renderer) drawTable(window draw.Window, current *game.Game, view viewport) {
	definition := current.Table
	// Contact geometry and sprite anchors come from the shared table definition.
	for _, lane := range definition.RolloverLanes {
		midpoint := lane.Segment.A.Add(lane.Segment.B).Mul(.5)
		path := "assets/images/lane-light-off.png"
		if current.LaneLit(lane.ID) {
			path = "assets/images/lane-light.png"
		}
		r.spriteCentered(window, path, view, midpoint, 34, 68, 0)
	}

	for _, bumper := range definition.Bumpers {
		placement := table.BumperMaterialFrame.Place(bumper.Center, bumper.Radius/table.BumperMaterialFrame.ContactRadius, 0)
		r.placedSprite(window, "assets/images/bumper-material.png", view, placement)
		r.placedSprite(window, "assets/images/bumper-emission.png", view, placement)
	}
	for _, post := range definition.Posts {
		r.placedSprite(window, "assets/images/post.png", view, table.PostFrame.Place(post.Center, post.Radius/table.PostFrame.ContactRadius, 0))
	}
	for _, target := range definition.DropTargets {
		path := "assets/images/target.png"
		if current.TargetDown(target.ID) {
			path = "assets/images/target-down.png"
		}
		midpoint := target.Segment.A.Add(target.Segment.B).Mul(.5)
		angle := math.Atan2(target.Segment.B.Y-target.Segment.A.Y, target.Segment.B.X-target.Segment.A.X) - math.Pi/2
		// The housing fits the capsule's complete contact extent.
		scale := (target.Segment.B.Sub(target.Segment.A).Length() + 2*target.Radius) / table.TargetFrame.ContactLength
		r.placedSprite(window, path, view, table.TargetFrame.Place(midpoint, scale, angle))
	}

	for _, flipper := range current.World.Flippers {
		scale := flipper.Length / table.FlipperFrame.Tip.Sub(table.FlipperFrame.Anchor).Length()
		r.placedSprite(window, "assets/images/flipper.png", view, table.FlipperFrame.Place(flipper.Pivot, scale, flipper.Angle))
	}

	// Keep the spring foot fixed inside the well while the head retracts.
	plungerY := definition.Plunger.Position.Y - 2 + current.PlungerCharge*8
	plungerHeight := 40 - current.PlungerCharge*16
	r.spriteCentered(window, "assets/images/plunger.png", view, physics.V(definition.Plunger.Position.X, plungerY), 36, plungerHeight, 0)
	if current.Ball.Active {
		r.placedSprite(window, "assets/images/ball.png", view, table.BallFrame.Place(current.Ball.Position, current.Ball.Radius/table.BallFrame.ContactRadius, 0))
	}
}

func (r *renderer) drawEffects(window draw.Window, view viewport, elapsed float64) {
	alive := r.particles[:0]
	for index, effect := range r.particles {
		effect.life -= elapsed
		if effect.life <= 0 {
			continue
		}
		progress := 1 - effect.life/effect.maxLife
		angle := float64(index)*2.399 + r.elapsedTime*.7
		distance := 58 * progress
		position := effect.position.Add(physics.V(math.Cos(angle)*distance, math.Sin(angle)*distance))
		size := 15 * (1 - progress*.65)
		tail := position.Sub(physics.V(math.Cos(angle)*size, math.Sin(angle)*size))
		r.thickLine(window, view, physics.Segment{A: tail, B: position}, draw.RGBA(.3, .94, 1, float32(1-progress)), 1)
		if index%5 == 0 {
			x, y := view.point(effect.position)
			radius := view.size(20 + progress*40)
			r.thickEllipse(window, x-radius, y-radius, radius*2, radius*2, draw.RGBA(.13, .91, 1, float32((1-progress)*.45)), view.stroke(1))
		}
		r.spriteCentered(window, "assets/images/particle.png", view, position, size, size, int(angle*180/math.Pi))
		alive = append(alive, effect)
	}
	r.particles = alive
}

func (r *renderer) drawHUD(window draw.Window, current *game.Game, view viewport) {
	scale := float32(view.scale)
	r.text(window, fmt.Sprintf("SCORE %08d", current.Score), view.x(55), view.y(17), scale*.9, cyan)
	r.text(window, fmt.Sprintf("BALL %d / 3", min(current.BallNumber, 3)), view.x(452), view.y(17), scale*.8, cyan)
	r.text(window, "HIGH", view.x(285), view.y(18), scale*.35, draw.RGB(.55, .65, .68))
	r.text(window, fmt.Sprintf("%08d", current.HighScore), view.x(285), view.y(31), scale*.35, draw.RGB(.55, .65, .68))
	r.text(window, "BONUS", view.x(597), view.y(18), scale*.35, amber)
	r.text(window, fmt.Sprintf("%d x%d", current.Bonus, current.BonusMultiplier), view.x(597), view.y(31), scale*.35, amber)
	r.centerText(window, "CIRCUIT COMPLETE", view.x(510), view.y(663), scale*.42, magenta)
	for i, lane := range current.Table.RolloverLanes {
		center := lane.Segment.A.Add(lane.Segment.B).Mul(.5)
		r.centerText(window, fmt.Sprintf("0%d", i+1), view.x(center.X), view.y(center.Y+48), scale*.42, cyan)
	}
}

func (r *renderer) drawState(window draw.Window, current *game.Game, view viewport) {
	centerX := view.offsetX + view.width/2
	centerY := view.offsetY + view.height/2
	switch current.State {
	case game.Loading:
		r.centerText(window, "LOADING RELAY...", centerX, centerY, float32(math.Max(1, view.scale*1.3)), cyan)
	case game.Attract:
		r.centerText(window, "PRESS ENTER TO CONNECT", centerX, view.y(711), float32(view.scale*.7), cyan)
		r.centerText(window, "A / LEFT     D / RIGHT", centerX, view.y(741), float32(view.scale*.48), draw.LightGray)
	case game.BallReady:
		r.centerText(window, "HOLD SPACE / DOWN TO CHARGE", centerX, view.y(720), float32(view.scale*.62), amber)
		barWidth := view.size(260)
		barHeight := view.size(16)
		strokeWidth := view.stroke(tableOutlineWidth)
		inset := (strokeWidth + 1) / 2
		r.thickRect(window, centerX-barWidth/2, view.y(750), barWidth, barHeight, draw.White, strokeWidth)
		fillWidth := int(float64(max(0, barWidth-2*inset)) * current.PlungerCharge)
		window.FillRect(centerX-barWidth/2+inset, view.y(750)+inset, fillWidth, max(1, barHeight-2*inset), magenta)
	case game.Paused:
		window.FillRect(view.offsetX, view.offsetY, view.width, view.height, draw.RGBA(0, 0, 0, .68))
		r.centerText(window, "PAUSED", centerX, centerY, float32(math.Max(1.2, view.scale*1.6)), amber)
		r.centerText(window, "PRESS P TO RESUME", centerX, centerY+view.size(54), float32(math.Max(.7, view.scale*.8)), draw.White)
	case game.BallLost:
		r.centerText(window, "BALL LOST", centerX, centerY, float32(math.Max(1, view.scale*1.35)), red)
	case game.GameOver:
		window.FillRect(view.offsetX, view.offsetY, view.width, view.height, draw.RGBA(0, 0, 0, .58))
		r.centerText(window, "GAME OVER", centerX, centerY-view.size(35), float32(math.Max(1.2, view.scale*1.6)), magenta)
		r.centerText(window, "PRESS ENTER TO RESTART", centerX, centerY+view.size(36), float32(math.Max(.75, view.scale*.85)), lime)
	case game.Playing:
	}
}

func (r *renderer) thickLine(window draw.Window, view viewport, segment physics.Segment, color draw.Color, width int) {
	ax, ay := view.point(segment.A)
	bx, by := view.point(segment.B)
	dx, dy := float64(bx-ax), float64(by-ay)
	length := math.Hypot(dx, dy)
	if length == 0 {
		return
	}
	width = view.stroke(width)
	if drawCanvasLine(ax, ay, bx, by, color, width) {
		return
	}
	normalX, normalY := -dy/length, dx/length
	for offset := -width / 2; offset <= width/2; offset++ {
		x := int(math.Round(normalX * float64(offset)))
		y := int(math.Round(normalY * float64(offset)))
		window.DrawLine(ax+x, ay+y, bx+x, by+y, color)
	}
}

func (r *renderer) thickEllipse(window draw.Window, x, y, width, height int, color draw.Color, strokeWidth int) {
	strokeWidth = oddStrokeWidth(strokeWidth)
	if drawCanvasEllipse(x, y, width, height, color, strokeWidth) {
		return
	}
	for offset := -strokeWidth / 2; offset <= strokeWidth/2; offset++ {
		window.DrawEllipse(x-offset, y-offset, width+2*offset, height+2*offset, color)
	}
}

func (r *renderer) thickRect(window draw.Window, x, y, width, height int, color draw.Color, strokeWidth int) {
	strokeWidth = oddStrokeWidth(strokeWidth)
	if drawCanvasRect(x, y, width, height, color, strokeWidth) {
		return
	}
	for offset := -strokeWidth / 2; offset <= strokeWidth/2; offset++ {
		window.DrawRect(x-offset, y-offset, width+2*offset, height+2*offset, color)
	}
}

func oddStrokeWidth(width int) int {
	width = max(1, width)
	if width%2 == 0 {
		width++
	}
	return width
}

func (r *renderer) spriteCentered(window draw.Window, path string, view viewport, center physics.Vec, width, height float64, rotation int) {
	w, h := view.size(width), view.size(height)
	x, y := view.point(center)
	r.image(window, path, x-w/2, y-h/2, w, h, rotation)
}

func (r *renderer) image(window draw.Window, path string, x, y, width, height, rotation int) bool {
	err := window.DrawImageFileTo(path, x, y, max(1, width), max(1, height), rotation)
	if err == nil {
		return true
	}
	if !errors.Is(err, draw.ErrImageLoading) && r.loadError == nil {
		r.loadError = fmt.Errorf("draw %s: %w", path, err)
	}
	return false
}

func (r *renderer) centerText(window draw.Window, text string, centerX, y int, scale float32, color draw.Color) {
	if drawCanvasText(text, centerX, y, scale, color, true) {
		return
	}
	width, _ := window.GetScaledTextSize(text, scale)
	window.DrawScaledText(text, centerX-width/2, y, scale, color)
}

func truncate(value string, limit int) string {
	value = strings.ReplaceAll(value, "\n", " ")
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	if limit <= 0 {
		return ""
	}
	if limit <= 3 {
		return strings.Repeat(".", limit)
	}
	return string(runes[:limit-3]) + "..."
}

func (r *renderer) text(window draw.Window, text string, x, y int, scale float32, color draw.Color) {
	if !drawCanvasText(text, x, y, scale, color, false) {
		window.DrawScaledText(text, x, y, scale, color)
	}
}

func (r *renderer) placedSprite(window draw.Window, path string, view viewport, p table.Placement) {
	r.spriteCentered(window, path, view, p.Center, p.Width, p.Height, int(math.Round(p.Angle)))
}
