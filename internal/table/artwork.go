package table

import (
	"math"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
)

// SpriteFrame records source-pixel bounds and the physical anchor. Padding and
// baked shadows may extend beyond contact geometry; they are never colliders.
type SpriteFrame struct {
	Width, Height float64
	Anchor        physics.Vec
	// ContactRadius is the source-pixel radius of a circular contact surface.
	ContactRadius float64
	// ContactLength is the full source-pixel length of a target housing.
	ContactLength float64
	// Tip identifies the flipper's tip in source-pixel coordinates.
	Tip physics.Vec
}

// Placement is the destination rectangle center and size before rotation.
type Placement struct {
	Center               physics.Vec
	Width, Height, Angle float64
}

// Place keeps the authored anchor on the physical mechanism during rotation.
// Angle is supplied in radians; the returned draw angle is in degrees.
func (s SpriteFrame) Place(anchor physics.Vec, scale, angle float64) Placement {
	offset := physics.V(s.Width/2, s.Height/2).Sub(s.Anchor).Mul(scale)
	sin, cos := math.Sincos(angle)
	return Placement{
		Center: anchor.Add(physics.V(offset.X*cos-offset.Y*sin, offset.X*sin+offset.Y*cos)),
		Width:  s.Width * scale, Height: s.Height * scale, Angle: angle * 180 / math.Pi,
	}
}

// These source frames are shared by asset authoring and runtime placement.
var (
	BumperFrame  = SpriteFrame{Width: 128, Height: 128, Anchor: physics.V(64, 64), ContactRadius: 55}
	PostFrame    = SpriteFrame{Width: 96, Height: 96, Anchor: physics.V(48, 48), ContactRadius: 38}
	BallFrame    = SpriteFrame{Width: 64, Height: 64, Anchor: physics.V(32, 31), ContactRadius: 27}
	FlipperFrame = SpriteFrame{Width: 360, Height: 128, Anchor: physics.V(60, 64), Tip: physics.V(296, 64)}
	TargetFrame  = SpriteFrame{Width: 128, Height: 192, Anchor: physics.V(64, 92), ContactLength: 84 * 2}
)

// BumperMaterialFrame shares a 2x canvas across each bumper's patch, shadow,
// body, and emission. Only the body has height; its contact radius stays 54
// table units. The surrounding patch and light cannot occlude a moving ball.
var BumperMaterialFrame = SpriteFrame{Width: 384, Height: 384, Anchor: physics.V(192, 192), ContactRadius: 108}

// TitleOrigin is printed on the playfield, between the upper pair and lower
// bumper. It has no height and never occludes the ball.
var TitleOrigin = physics.V(256, 440)
