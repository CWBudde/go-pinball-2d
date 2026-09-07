package main

import (
	"image"
	"math"
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
)

// Opaque pixels represent physical hardware. Translucent penumbras may extend
// beyond it; moving the art to 2x must not silently enlarge contact surfaces.
func TestHardwareOpaqueFootprints(t *testing.T) {
	tests := []struct {
		name     string
		make     func() image.Image
		distance func(physics.Vec) float64
		radius   float64
	}{
		{"post", postImage, func(p physics.Vec) float64 { return p.Sub(table.PostFrame.Anchor).Length() }, table.PostFrame.ContactRadius},
		{"flipper", flipperImage, func(p physics.Vec) float64 {
			return physics.DistancePointSegment(p, physics.Segment{A: table.FlipperFrame.Anchor, B: table.FlipperFrame.Tip})
		}, 32 * 118.0 / 105},
		{"target", targetImage, func(p physics.Vec) float64 {
			r := 168 * 12 / (math.Sqrt(2000) + 24)
			return physics.DistancePointSegment(p, physics.Segment{A: physics.V(64, 8+r), B: physics.V(64, 176-r)})
		}, 168 * 12 / (math.Sqrt(2000) + 24)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := tt.make()
			bounds := img.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0xf000 && tt.distance(physics.V(float64(x)+.5, float64(y)+.5)) > tt.radius+1 {
						t.Fatalf("opaque hardware at (%d,%d) exceeds physical footprint", x, y)
					}
				}
			}
		})
	}
}
