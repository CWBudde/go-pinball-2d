package main

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/table"
)

func TestMaterialBodyFitsPhysicalFootprint(t *testing.T) {
	for name, img := range map[string]image.Image{"idle": bumperMaterialImage(), "compressed": bumperHitImage()} {
		t.Run(name, func(t *testing.T) {
			frame := table.BumperMaterialFrame
			for y := 0; y < img.Bounds().Dy(); y++ {
				for x := 0; x < img.Bounds().Dx(); x++ {
					_, _, _, alpha := img.At(x, y).RGBA()
					distance := math.Hypot(float64(x)+.5-frame.Anchor.X, float64(y)+.5-frame.Anchor.Y)
					// One source pixel of antialiasing is half a logical pixel. Dark
					// translucent fastener shadows are decorative, never contact edges.
					if alpha > 0xf000 && distance > frame.ContactRadius+1 {
						t.Fatalf("opaque body at (%d,%d) extends beyond contact radius: %.2f", x, y, distance)
					}
				}
			}
			for _, offset := range [][2]int{{106, 0}, {-107, 0}, {0, 106}, {0, -107}} {
				_, _, _, alpha := img.At(int(frame.Anchor.X)+offset[0], int(frame.Anchor.Y)+offset[1]).RGBA()
				if alpha < 0xf000 {
					t.Fatalf("visible contact edge is inset at %v", offset)
				}
			}
		})
	}
}

func TestMaterialLayersHaveTransparentPadding(t *testing.T) {
	for name, img := range map[string]image.Image{
		"body": bumperMaterialImage(), "patch": bumperPatchImage(),
		"shadow": bumperShadowImage(), "emission": bumperEmissionImage(),
	} {
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				if x >= 4 && y >= 4 && x < bounds.Max.X-4 && y < bounds.Max.Y-4 {
					continue
				}
				_, _, _, alpha := img.At(x, y).RGBA()
				if alpha != 0 {
					t.Fatalf("%s has clipped alpha at (%d,%d)", name, x, y)
				}
			}
		}
	}
}

func TestMaterialDownsamplePreservesTransparentEdgeColor(t *testing.T) {
	c := newCanvas(1, 1)
	// Hidden green RGB must not contaminate a single white coverage sample.
	for y := 0; y < supersample; y++ {
		for x := 0; x < supersample; x++ {
			c.img.SetNRGBA(x, y, color.NRGBA{G: 255})
		}
	}
	c.img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	got := c.finishAlpha().NRGBAAt(0, 0)
	want := color.NRGBA{R: 255, G: 255, B: 255, A: 28}
	if got != want {
		t.Fatalf("transparent-edge downsample = %v, want %v", got, want)
	}
}
