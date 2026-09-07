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
	instrumentFont,
	"assets/images/favicon.png",
	"assets/images/ball.png",
	"assets/images/ball-shadow.png",
	"assets/images/flipper.png",
	"assets/images/bumper.png",
	"assets/images/bumper-material.png",
	"assets/images/bumper-hit.png",
	"assets/images/bumper-patch.png",
	"assets/images/bumper-shadow.png",
	"assets/images/bumper-emission.png",
	"assets/images/post.png",
	"assets/images/target.png",
	"assets/images/target-down.png",
	"assets/images/lane-light.png",
	"assets/images/lane-light-off.png",
	"assets/images/plunger-head.png",
	"assets/images/plunger-coil.png",
	"assets/images/plunger-rod.png",
	"assets/images/glow.png",
	"assets/images/particle.png",
}

const tableOutlineWidth = 3

var (
	ink     = draw.RGB(.02, .03, .09)
	cyan    = draw.RGB(.13, .91, 1)
	magenta = draw.RGB(1, .17, .68)
	amber   = draw.RGB(1, .75, .22)
	red     = draw.RGB(1, .22, .3)
)

type renderer struct {
	loaded         map[string]bool
	loadError      error
	particles      []particle
	pulses         map[string]lightPulse
	targetLift     map[string]float64
	sequenceAge    float64
	sequenceActive bool
	elapsedTime    float64
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

func (r *renderer) draw(window draw.Window, current *game.Game, statusError error) {
	// Assets have antialiased source edges, but nearest-neighbor minification
	// discards that coverage. Let the backend filter them at the current size
	// (trilinear mipmaps on GLFW), including rotated and moving sprites.
	window.BlurImages(true)
	width, height := window.Size()
	window.FillRect(0, 0, width, height, ink)
	view := newViewport(width, height)

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
	r.drawLightPools(window, current, view)
	// Layer order: flat playfield, cast shadows, static mechanisms, dynamic
	// mechanisms/ball, safe foreground covers, emission/effects, instrument HUD.
	r.image(window, "assets/images/table-shadows.png", view.offsetX, view.offsetY, view.width, view.height, 0)
	r.image(window, "assets/images/table-hardware.png", view.offsetX, view.offsetY, view.width, view.height, 0)
	r.drawTableDisplays(window, current, view)
	r.drawTable(window, current, view)
	r.image(window, "assets/images/table-foreground.png", view.offsetX, view.offsetY, view.width, view.height, 0)
	r.drawEffects(window, view)
	r.drawHUD(window, current, view)

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
		r.laneResponse(window, view, lane, r.intensity(lane.ID))
	}

	for _, bumper := range definition.Bumpers {
		placement := table.BumperMaterialFrame.Place(bumper.Center, bumper.Radius/table.BumperMaterialFrame.ContactRadius, 0)
		path := "assets/images/bumper-material.png"
		compression := 0.0
		if pulse, ok := r.pulses[bumper.ID]; ok && pulse.age < .095 {
			path = "assets/images/bumper-hit.png"
			compression = 3
		}
		r.placedSprite(window, path, view, placement)
		placement.Center.Y += compression
		r.placedSprite(window, "assets/images/bumper-emission.png", view, placement)
		r.bumperResponse(window, view, bumper, compression, r.intensity(bumper.ID))
	}
	for _, post := range definition.Posts {
		r.placedSprite(window, "assets/images/post.png", view, table.PostFrame.Place(post.Center, post.Radius/table.PostFrame.ContactRadius, 0))
	}
	for _, sling := range definition.Slingshots {
		r.slingResponse(window, view, sling)
	}
	for _, target := range definition.DropTargets {
		r.drawTarget(window, current, view, target)
	}

	for _, flipper := range current.World.Flippers {
		scale := flipper.Length / table.FlipperFrame.Tip.Sub(table.FlipperFrame.Anchor).Length()
		r.placedSprite(window, "assets/images/flipper.png", view, table.FlipperFrame.Place(flipper.Pivot, scale, flipper.Angle))
	}

	r.drawPlunger(window, current, view)
	if current.Ball.Active {
		r.drawBall(window, current, view)
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
	// Alternate engine surfaces can provide an antialiased thick stroke without
	// the integer-offset fallback used by older window backends.
	if surface, ok := window.(interface {
		StrokeLine(int, int, int, int, int, draw.Color)
	}); ok {
		surface.StrokeLine(ax, ay, bx, by, width, color)
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
