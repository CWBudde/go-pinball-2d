package main

import (
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
)

// Mechanism sprites use 2x source pixels, with the same upper-left key as the
// bumper. Author in logical sprite coordinates and filter premultiplied alpha.
func materialCanvas(w, h int) *canvas {
	c := newCanvas(w*2, h*2)
	c.s *= 2
	return c
}

// capsuleSurface shades a constant-radius contact surface. The distance field
// keeps rounded ends and long faces within the physical capsule, even on flippers.
func (c *canvas) capsuleSurface(ax, ay, bx, by, radius float64, shade func(x, y, r float64) (color.NRGBA, float64)) {
	cx, cy := (ax+bx)/2, (ay+by)/2
	hx, hy := math.Abs(bx-ax)/2+radius, math.Abs(by-ay)/2+radius
	for py := int((cy - hy - 1) * c.s); py <= int((cy+hy+1)*c.s); py++ {
		for px := int((cx - hx - 1) * c.s); px <= int((cx+hx+1)*c.s); px++ {
			x, y := (float64(px)+.5)/c.s, (float64(py)+.5)/c.s
			d := pointSegmentDistance(x, y, ax, ay, bx, by)
			coverage := clamp((radius-d)*c.s+.5, 0, 1)
			if coverage > 0 {
				col, alpha := shade((x-cx)/hx, (y-cy)/hy, d/radius)
				c.blendPixel(px, py, col, alpha*coverage)
			}
		}
	}
}

func rubberMaterial(x, y, r float64) (color.NRGBA, float64) {
	v := 9 + 12*math.Max(0, -.5*x-.8*y) + 2*math.Sin(r*90)
	return materialRGB(v*1.05, v, v*1.1), 1
}

func glassMaterial(x, y, r float64) (color.NRGBA, float64) {
	v := 8 + 14*bell(x+y, -.65, .6) + 20*bell(y+.5*x, -.55, .055)*(1-r)
	return materialRGB(v, v*1.25, v*1.45), 1
}

func postImage() image.Image {
	c := materialCanvas(48, 48)
	c.circle(25, 28, 20, black, .5)
	c.ellipseSurface(24, 24, 19, 19, bumperChrome)
	c.ellipseSurface(24, 24, 17.5, 17.5, rubberMaterial)
	c.ellipseSurface(24, 25, 14.5, 14.5, bumperChrome)
	c.ellipseSurface(24, 22, 14, 12, bumperChrome)
	c.ellipseSurface(24, 22, 11.5, 9.5, glassMaterial)
	c.ellipseArc(24, 22, 10, 8, math.Pi, math.Pi*2, 1.8, cyan, .9)
	c.bumperFastener(24, 22)
	return c.finishAlpha()
}

func flipperImage() image.Image {
	c := materialCanvas(180, 64)
	// The physical radius is 16 at scale 105/118. Keep the whole blade and hub
	// inside that capsule; only the translucent shadow extends below it.
	const r = 16.0 * 118 / 105
	c.line(32, 37, 150, 37, r*2+5, black, .45)
	c.capsuleSurface(30, 32, 148, 32, r, rubberMaterial)
	c.capsuleSurface(30, 32, 148, 32, r-1.4, func(x, y, d float64) (color.NRGBA, float64) {
		v := .3 + .5*math.Max(0, -y) + .12*bell(d, .8, .08)
		return materialRGB(220*v, 27*v, 79*v), 1
	})
	c.capsuleSurface(30, 30, 148, 30, r-4, bumperChrome)
	c.capsuleSurface(30, 29, 148, 29, r-5.5, func(x, y, d float64) (color.NRGBA, float64) {
		v := 185 + 44*bell(y, -.5, .5) - 48*math.Max(0, y) + 2*math.Sin(x*460)
		return materialRGB(v, v*.99, v*.96), 1
	})
	c.line(52, 21, 137, 21, 1.1, cyanWhite, .7)
	c.line(55, 37, 137, 37, .6, metal, .5)
	for i := 0; i < 3; i++ {
		x := 61 + float64(i)*5
		c.line(x, 26, x+2, 29, .6, metal, .7)
	}
	c.ellipseSurface(30, 32, 16.7, 16.7, bumperChrome)
	c.ellipseSurface(30, 31, 13, 13, rubberMaterial)
	c.ellipseSurface(30, 30, 10.5, 10.5, bumperChrome)
	c.ellipseSurface(30, 30, 8.2, 8.2, func(x, y, r float64) (color.NRGBA, float64) {
		v := 135 + 90*math.Max(0, -.4*x-.7*y) - 45*r
		return materialRGB(v, v*1.02, v*1.04), 1
	})
	return c.finishAlpha()
}

func targetImage() image.Image     { return targetFace(false) }
func targetDownImage() image.Image { return targetFace(true) }
func targetFace(down bool) image.Image {
	c := materialCanvas(64, 96)
	// Segment and cap radii have exactly the same ratio as the table collider.
	radius := 84 * 12 / (math.Sqrt(2000) + 24)
	a, b := 4+radius, 88-radius
	if down {
		// A dropped target is a flush socket. The ball can cross it without
		// appearing to pass through a still-raised frame or mounting screws.
		c.capsuleSurface(32, a, 32, b, radius, func(x, y, r float64) (color.NRGBA, float64) {
			v := 8 + 16*bell(r, .95, .025)
			return materialRGB(v, v*1.13, v*1.25), 1
		})
		c.line(32, 18, 32, 74, 7, black, 1)
		for y := 24.0; y < 72; y += 10 {
			c.line(28, y, 36, y, 1, metal, .25)
		}
		return c.finishAlpha()
	}
	c.line(34, a+4, 34, b+4, radius*2+5, black, .5)
	c.capsuleSurface(32, a, 32, b, radius, rubberMaterial)
	c.capsuleSurface(32, a, 32, b, radius-1.4, bumperChrome)
	c.capsuleSurface(32, a, 32, b, radius-3.2, func(x, y, r float64) (color.NRGBA, float64) { return black, 1 })
	c.capsuleSurface(32, a+2, 32, b-2, radius-4.2, glassMaterial)
	c.line(23, 22, 23, 69, .8, silver, .55)
	c.line(41, 22, 41, 69, 1, pink, .5)
	for i, length := range []float64{7, 14, 20, 10} {
		y := 30 + float64(i)*10
		c.glowLine(32-length/2, y, 32+length/2, y, 2.4, pink, .9)
		c.line(32-length/2, y, 32+length/2, y, .65, cyanWhite, .9)
	}
	c.screw(32, 12, 2.7)
	c.screw(32, 80, 2.7)
	return c.finishAlpha()
}

func laneLightImage() image.Image    { return laneLight(true) }
func laneLightOffImage() image.Image { return laneLight(false) }
func laneLight(lit bool) image.Image {
	c := materialCanvas(48, 96)
	c.capsuleSurface(24, 20, 24, 76, 16, bumperChrome)
	c.capsuleSurface(24, 20, 24, 76, 14, func(x, y, r float64) (color.NRGBA, float64) { return black, 1 })
	c.capsuleSurface(24, 20, 24, 76, 12, glassMaterial)
	col, opacity := metal, .5
	if lit {
		col, opacity = cyan, 1
	}
	for _, y := range []float64{28, 47, 66} {
		c.glowLine(17, y+5, 24, y-3, 2.5, col, opacity)
		c.glowLine(24, y-3, 31, y+5, 2.5, col, opacity)
	}
	c.line(15, 22, 15, 65, .7, silver, .3)
	c.circle(24, 81, 3.2, black, 1)
	c.circle(24, 81, 1.8, gold, opacity)
	return c.finishAlpha()
}

func plungerImage() image.Image {
	c := materialCanvas(56, 180)
	c.capsuleSurface(28, 13, 28, 156, 6, bumperChrome)
	// Helical wire wraps a polished shaft, with dark rear turns and a narrow key.
	for y := 43.0; y < 143; y += 9 {
		c.line(17, y-1, 39, y+5, 5, black, 1)
		c.line(17, y+5, 39, y-1, 5, graphite, 1)
		c.line(17, y+4, 39, y-2, 3.2, metal, 1)
		c.line(18, y+3, 38, y-3, 1.1, silver, 1)
	}
	c.capsuleSurface(20, 11, 36, 11, 7, bumperChrome)
	c.capsuleSurface(20, 151, 36, 151, 8, bumperChrome)
	c.capsuleSurface(20, 163, 36, 163, 10, rubberMaterial)
	c.capsuleSurface(20, 161, 36, 161, 8, func(x, y, r float64) (color.NRGBA, float64) {
		v := .4 + .5*math.Max(0, -y)
		return materialRGB(255*v, 48*v, 129*v), 1
	})
	c.line(21, 156, 35, 156, 1, silver, .75)
	return c.finishAlpha()
}

// Chrome guides retain the exact shared centerline and contact width. All
// highlights lie inside that width; broader shapes belong only to shadows.
func (c *canvas) rails(walls []physics.LineCollider) {
	for _, layer := range []struct {
		inset, dx, dy float64
		col           color.NRGBA
	}{
		{0, 0, 0, black},
		{1, 0, 0, metal},
		{2.7, -.4, -.6, silver},
		{4.3, .5, .7, black},
		{6, 0, 0, graphite},
	} {
		for _, w := range walls {
			a, b := w.Segment.A, w.Segment.B
			c.line(a.X+layer.dx, a.Y+layer.dy, b.X+layer.dx, b.Y+layer.dy, math.Max(1, w.Radius*2-layer.inset), layer.col, 1)
		}
	}
	// Sparse captive brackets break long runs into assembled sections. They fit
	// inside the existing contact face, so no uncollided protrusions are introduced.
	for _, w := range walls {
		delta := w.Segment.B.Sub(w.Segment.A)
		length := delta.Length()
		if length < 48 {
			if strings.HasPrefix(w.ID, "guide_arch_") && (strings.HasSuffix(w.ID, "_01") || strings.HasSuffix(w.ID, "_06") || strings.HasSuffix(w.ID, "_12")) {
				p := w.Segment.A.Add(w.Segment.B).Mul(.5)
				c.ellipseSurface(p.X, p.Y, 6.5, 6.5, bumperChrome)
				c.bumperFastener(p.X, p.Y)
			}
			continue
		}
		steps := int(length/135) + 1
		for i := 0; i < steps; i++ {
			t := (float64(i) + .5) / float64(steps)
			p := w.Segment.A.Add(delta.Mul(t))
			c.ellipseSurface(p.X, p.Y, w.Radius-.4, w.Radius-.4, bumperChrome)
			c.line(p.X-1, p.Y-1, p.X+1, p.Y+1, .8, black, 1)
		}
	}
}

// Flush routed pads and panel seams tie the four sockets into a target bank.
// None of this ink has height: dropped targets and passing balls stay readable.
func (c *canvas) boardDetails() {
	for i := 0; i < 4; i++ {
		x, y := 568-float64(i)*13, 424+float64(i)*62
		c.line(x+5, y-20, x+5, y+20, 1.2, black, 1)
		c.line(x+6, y-20, x+6, y+20, .6, silver, .35)
		c.circle(x, y, 3, black, 1)
		c.ring(x, y, 2.6, 1.6, metal, .8)
		for j := 0; j < 3; j++ {
			xx := x + 10 + float64(j)*3
			c.line(xx, y-12, xx, y+10, .6, metal, .4)
		}
		if i < 3 {
			c.line(x+5, y+20, x-8, y+42, .8, pink, .3)
		}
	}
	// Small signal pads end at assemblies rather than becoming floating dots.
	for _, p := range [][2]float64{{154, 610}, {561, 610}, {150, 295}, {545, 286}, {283, 695}, {403, 699}} {
		for i := 0; i < 4; i++ {
			x := p[0] + float64(i)*4
			c.line(x, p[1], x, p[1]+7, 1, metal, .65)
		}
	}
	// Central etched seams leave both return and drain channels visually clear.
	for _, p := range [][][2]float64{
		{{240, 698}, {278, 736}, {278, 857}, {325, 904}},
		{{388, 707}, {368, 727}, {368, 857}, {338, 887}},
	} {
		for i := 1; i < len(p); i++ {
			a, b := p[i-1], p[i]
			c.line(a[0]+1, a[1]+1, b[0]+1, b[1]+1, 1.5, black, .8)
			c.line(a[0], a[1], b[0], b[1], .6, metal, .45)
		}
	}
}

// Separate moving plunger parts retain their metal finish during compression.
func plungerHeadImage() image.Image {
	c := materialCanvas(36, 12)
	c.capsuleSurface(8, 6, 28, 6, 4, bumperChrome)
	c.line(9, 8, 27, 8, .7, pink, .75)
	return c.finishAlpha()
}

func plungerCoilImage() image.Image {
	c := materialCanvas(32, 8)
	c.line(5, 2, 27, 6, 2.8, black, 1)
	c.line(5, 6, 27, 2, 2.8, metal, 1)
	c.line(5, 5.5, 27, 1.5, .9, silver, 1)
	return c.finishAlpha()
}

func plungerRodImage() image.Image {
	c := materialCanvas(6, 32)
	c.capsuleSurface(3, 3, 3, 29, 2, bumperChrome)
	return c.finishAlpha()
}
