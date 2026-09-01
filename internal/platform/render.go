package platform

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/gonutz/prototype/draw"
)

var requiredImages = []string{
	"assets/images/background.png",
	"assets/images/logo.png",
	"assets/images/favicon.png",
	"assets/images/ball.png",
	"assets/images/flipper.png",
	"assets/images/bumper.png",
	"assets/images/post.png",
	"assets/images/target.png",
	"assets/images/lane-light.png",
	"assets/images/plunger.png",
	"assets/images/glow.png",
	"assets/images/particle.png",
}

const (
	tableOutlineWidth       = 3
	flipperAssetWidth       = 180.0
	flipperAssetHeight      = 64.0
	flipperAssetPivotX      = 30.0
	flipperAssetTipX        = 148.0
	flipperAssetAnchorWidth = flipperAssetTipX - flipperAssetPivotX
)

var (
	ink     = draw.RGB(.02, .03, .09)
	cyan    = draw.RGB(.13, .91, 1)
	magenta = draw.RGB(1, .17, .68)
	violet  = draw.RGB(.49, .28, 1)
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
	r.drawTable(window, current, view)
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
		window.DrawScaledText(message, view.offsetX+view.size(14), view.offsetY+view.height-view.size(38), float32(math.Max(.65, view.scale*.8)), red)
	}
}

func (r *renderer) drawTable(window draw.Window, current *game.Game, view viewport) {
	definition := current.Table
	for _, line := range append(append(append([]physics.LineCollider{}, definition.OuterWalls...), definition.ShooterLane...), definition.GuideWalls...) {
		r.thickLine(window, view, line.Segment, cyan, tableOutlineWidth)
	}
	for _, sling := range definition.Slingshots {
		r.thickLine(window, view, physics.Segment{A: sling.Triangle[0], B: sling.Triangle[1]}, magenta, tableOutlineWidth)
		r.thickLine(window, view, physics.Segment{A: sling.Triangle[1], B: sling.Triangle[2]}, magenta, tableOutlineWidth)
		r.thickLine(window, view, physics.Segment{A: sling.Triangle[2], B: sling.Triangle[0]}, magenta, tableOutlineWidth)
	}

	for _, lane := range definition.RolloverLanes {
		midpoint := lane.Segment.A.Add(lane.Segment.B).Mul(.5)
		color := violet
		if current.LaneLit(lane.ID) {
			color = lime
		}
		r.spriteCentered(window, "assets/images/lane-light.png", view, midpoint, 28, 56, 90)
		x, y := view.point(midpoint)
		r.thickEllipse(window, x-view.size(13), y-view.size(6), view.size(26), view.size(12), color, tableOutlineWidth)
	}

	for _, bumper := range definition.Bumpers {
		r.spriteCentered(window, "assets/images/bumper.png", view, bumper.Center, bumper.Radius*2.3, bumper.Radius*2.3, 0)
	}
	for _, post := range definition.Posts {
		r.spriteCentered(window, "assets/images/post.png", view, post.Center, post.Radius*2.4, post.Radius*2.4, 0)
	}
	for _, target := range definition.DropTargets {
		if current.TargetDown(target.ID) {
			continue
		}
		midpoint := target.Segment.A.Add(target.Segment.B).Mul(.5)
		angle := math.Atan2(target.Segment.B.Y-target.Segment.A.Y, target.Segment.B.X-target.Segment.A.X)*180/math.Pi - 90
		r.spriteCentered(window, "assets/images/target.png", view, midpoint, 32, 48, int(math.Round(angle)))
	}

	for _, flipper := range current.World.Flippers {
		midpoint := flipper.Pivot.Add(flipper.Tip()).Mul(.5)
		angle := int(math.Round(flipper.Angle * 180 / math.Pi))
		scale := flipper.Length / flipperAssetAnchorWidth
		r.spriteCentered(window, "assets/images/flipper.png", view, midpoint, flipperAssetWidth*scale, flipperAssetHeight*scale, angle)
	}

	plungerY := definition.Plunger.Position.Y - 52 + current.PlungerCharge*22
	r.spriteCentered(window, "assets/images/plunger.png", view, physics.V(definition.Plunger.Position.X, plungerY), 36, 116, 0)
	if current.Ball.Active {
		r.spriteCentered(window, "assets/images/glow.png", view, current.Ball.Position, 58, 58, 0)
		r.spriteCentered(window, "assets/images/ball.png", view, current.Ball.Position, current.Ball.Radius*2, current.Ball.Radius*2, 0)
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
		r.spriteCentered(window, "assets/images/particle.png", view, position, size, size, int(angle*180/math.Pi))
		alive = append(alive, effect)
	}
	r.particles = alive
}

func (r *renderer) drawHUD(window draw.Window, current *game.Game, view viewport) {
	scale := float32(math.Max(.72, view.scale*.92))
	x := view.offsetX + view.size(24)
	y := view.offsetY + view.size(18)
	window.DrawScaledText(fmt.Sprintf("SCORE %08d", current.Score), x, y, scale, draw.White)
	window.DrawScaledText(fmt.Sprintf("HIGH  %08d", current.HighScore), x, y+view.size(27), scale*.72, cyan)
	right := view.offsetX + view.width - view.size(205)
	window.DrawScaledText(fmt.Sprintf("BALL %d/3", min(current.BallNumber, 3)), right, y, scale*.78, amber)
	window.DrawScaledText(fmt.Sprintf("BONUS %d x%d", current.Bonus, current.BonusMultiplier), right, y+view.size(27), scale*.68, lime)
}

func (r *renderer) drawState(window draw.Window, current *game.Game, view viewport) {
	centerX := view.offsetX + view.width/2
	centerY := view.offsetY + view.height/2
	switch current.State {
	case game.Loading:
		r.centerText(window, "LOADING RELAY...", centerX, centerY, float32(math.Max(1, view.scale*1.3)), cyan)
	case game.Attract:
		r.image(window, "assets/images/logo.png", view.x(80), view.y(245), view.size(560), view.size(175), 0)
		r.centerText(window, "PRESS ENTER TO START", centerX, view.y(500), float32(math.Max(.9, view.scale)), lime)
		r.centerText(window, "A / LEFT     D / RIGHT     SPACE / DOWN", centerX, view.y(550), float32(math.Max(.62, view.scale*.7)), draw.LightGray)
	case game.BallReady:
		r.centerText(window, "HOLD SPACE / DOWN TO CHARGE", centerX, view.y(720), float32(math.Max(.72, view.scale*.78)), amber)
		barWidth := view.size(260)
		r.thickRect(window, centerX-barWidth/2, view.y(750), barWidth, view.size(16), draw.White, tableOutlineWidth)
		window.FillRect(centerX-barWidth/2+2, view.y(750)+2, int(float64(barWidth-4)*current.PlungerCharge), max(1, view.size(16)-4), magenta)
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
	width = oddStrokeWidth(width)
	normalX, normalY := -dy/length, dx/length
	for offset := -width / 2; offset <= width/2; offset++ {
		x := int(math.Round(normalX * float64(offset)))
		y := int(math.Round(normalY * float64(offset)))
		window.DrawLine(ax+x, ay+y, bx+x, by+y, color)
	}
}

func (r *renderer) thickEllipse(window draw.Window, x, y, width, height int, color draw.Color, strokeWidth int) {
	strokeWidth = oddStrokeWidth(strokeWidth)
	for offset := -strokeWidth / 2; offset <= strokeWidth/2; offset++ {
		window.DrawEllipse(x-offset, y-offset, width+2*offset, height+2*offset, color)
	}
}

func (r *renderer) thickRect(window draw.Window, x, y, width, height int, color draw.Color, strokeWidth int) {
	strokeWidth = oddStrokeWidth(strokeWidth)
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
	width, _ := window.GetScaledTextSize(text, scale)
	window.DrawScaledText(text, centerX-width/2, y, scale, color)
}

func truncate(value string, limit int) string {
	value = strings.ReplaceAll(value, "\n", " ")
	if len(value) <= limit {
		return value
	}
	return value[:limit-3] + "..."
}
