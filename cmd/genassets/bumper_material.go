package main

import (
	"image"
	"image/color"
	"math"

	"github.com/CWBudde/go-pinball-2d/internal/table"
)

// Shared bumper material: one cool softbox above/left, cyan LEDs at the upper left,
// pink LEDs around the remaining cap. Work in table pixels, output at 2x,
// supersample at 6x. Material shading is baked; runtime adds timed emission
// and selects a compressed cap variant after an impact.
func newBumperCanvas() *canvas {
	c := newCanvas(int(table.BumperMaterialFrame.Width), int(table.BumperMaterialFrame.Height))
	c.s *= 2
	return c
}

// ellipseSurface supplies normalized coordinates to an authored material shader.
func (c *canvas) ellipseSurface(cx, cy, rx, ry float64, shade func(x, y, r float64) (color.NRGBA, float64)) {
	for py := int((cy - ry - 1) * c.s); py <= int((cy+ry+1)*c.s); py++ {
		for px := int((cx - rx - 1) * c.s); px <= int((cx+rx+1)*c.s); px++ {
			x, y := ((float64(px)+.5)/c.s-cx)/rx, ((float64(py)+.5)/c.s-cy)/ry
			r := math.Hypot(x, y)
			coverage := clamp((1-r)*math.Min(rx, ry)*c.s+.5, 0, 1)
			if coverage > 0 {
				col, opacity := shade(x, y, r)
				c.blendPixel(px, py, col, opacity*coverage)
			}
		}
	}
}

func bell(x, center, width float64) float64 {
	t := (x - center) / width
	return math.Exp(-t * t)
}

func materialRGB(r, g, b float64) color.NRGBA {
	return color.NRGBA{R: uint8(clamp(r, 0, 255) + .5), G: uint8(clamp(g, 0, 255) + .5), B: uint8(clamp(b, 0, 255) + .5), A: 255}
}

// Brushed chrome reflects a broad key and thin softbox edges, separated by dark
// bands. The directional changes distinguish metal from a uniformly glowing rim.
func bumperChrome(x, y, r float64) (color.NRGBA, float64) {
	a := math.Atan2(y, x)
	key := math.Max(0, -.62*x-.78*y)
	bands := 105*bell(key, .77, .14) + 115*bell(key, .97, .045)
	bevel := 24*bell(r, .982, .013) - 32*bell(r, .947, .017)
	grain := 2 * math.Sin(r*2100+a*2)
	v := 25 + 38*key + bands + bevel + grain
	cyanReflection := 30 * bell(a, -2.6, .35)
	pinkReflection := 40 * bell(a, .9, .5)
	return materialRGB(v+pinkReflection, v*1.04+cyanReflection, v*1.09+cyanReflection+pinkReflection*.4), 1
}

func bumperMaterialImage() image.Image { return bumperBody(0) }
func bumperHitImage() image.Image      { return bumperBody(3) }

func bumperBody(compression float64) image.Image {
	c := newBumperCanvas()
	// The complete opaque footprint stays inside the existing radius-54 collider.
	c.ellipseSurface(96, 96, 54, 54, bumperChrome)
	c.ellipseSurface(96, 96, 52.3, 52.3, func(x, y, r float64) (color.NRGBA, float64) {
		// Matte rubber: broad diffuse response and shallow concentric molding ribs.
		v := 9 + 10*math.Max(0, -.6*x-.8*y) + 3*math.Sin(r*240)
		return materialRGB(v, v*1.1, v*1.2), 1
	})
	// An extruded cylinder sits on the rubber skirt. The lower ellipse remains
	// visible beneath the raised cap, with vertical metal reflections and seams.
	c.ellipseSurface(96, 101, 47.5, 44, func(x, y, r float64) (color.NRGBA, float64) {
		v := 13 + 115*bell(x, -.69, .075) + 52*bell(x, .78, .045) + 15*bell(x, -.4, .25)
		v *= .75 + .25*y
		v += 35 * bell(r, .975, .015)
		seam := math.Pow(math.Max(0, math.Cos(x*36)), 30)
		v *= 1 - seam*.65
		p := 48 * bell(x, .4, .38) * math.Max(0, y)
		return materialRGB(v+p, v*1.02+10*bell(x, -.8, .2), v*1.1+p*.3), 1
	})
	c.ellipseSurface(96, 89+compression, 48, 44, bumperChrome)
	c.ellipseSurface(96, 89+compression, 45.9, 41.9, func(_, _, _ float64) (color.NRGBA, float64) { return black, 1 })
	// Deep LED channel, with a narrow reflective inner lip.
	c.ellipseSurface(96, 89+compression, 43.8, 39.8, func(x, y, r float64) (color.NRGBA, float64) {
		col, _ := bumperChrome(x, y, r)
		return mix(col, pink, .33), 1
	})
	c.ellipseSurface(96, 89+compression, 39.8, 35.8, func(_, _, _ float64) (color.NRGBA, float64) { return black, 1 })
	c.ellipseSurface(96, 89+compression, 37.8, 33.8, bumperChrome)
	// Smoked glass: recessed edges, cool upper-left reflection, fine substrate
	// grain and a restrained diagonal softbox streak. No uniform neon outline.
	c.ellipseSurface(96, 89+compression, 36.4, 32.4, func(x, y, r float64) (color.NRGBA, float64) {
		key := bell(x, -.32, .7) * bell(y, -.48, .65)
		grain := 1.2 * math.Sin(x*440+y*217) * math.Sin(y*381-x*59)
		v := 8 + 19*key + grain
		edge := math.Pow(clamp(r, 0, 1), 10)
		streak := bell(y+.45*x, -.57, .065) * (1 - edge)
		return materialRGB(v+20*streak+12*edge*math.Max(x, 0), v*1.3+29*streak+7*edge, v*1.55+32*streak+11*edge), 1
	})
	// Tiny routed pads under the glass, followed by an inset relay schematic.
	for _, x := range []float64{77, 82, 87, 92} {
		c.line(x, 108, x+3, 108, .5, silver, .25)
	}
	c.line(83, 67, 100, 67, .6, metal, .6)
	c.line(100, 67, 109, 76, .6, metal, .6)
	c.relay(96, 92+compression, .69, cyan)
	// Three captive screws are contained by the skirt, not floating outside it.
	for _, a := range []float64{math.Pi / 2, math.Pi * 7 / 6, math.Pi * 11 / 6} {
		c.bumperFastener(96+math.Cos(a)*48, 96+math.Sin(a)*48)
	}
	return c.finishAlpha()
}

func (c *canvas) bumperFastener(x, y float64) {
	c.circle(x+.5, y+1, 5.6, black, .9)
	c.ellipseSurface(x, y, 4.8, 4.8, bumperChrome)
	c.ellipseSurface(x, y, 3.7, 3.7, func(nx, ny, r float64) (color.NRGBA, float64) {
		v := 60 + 140*math.Max(0, -.5*nx-.7*ny) + 30*(1-r)
		return materialRGB(v, v*1.02, v*1.04), 1
	})
	c.line(x-1.8, y+1.8, x+1.8, y-1.8, 1.2, black, 1)
	c.line(x-1.3, y-1.3, x+1.3, y+1.3, .9, black, 1)
	c.line(x-1.5, y+2.1, x+.2, y+.4, .45, silver, .85)
}

func (c *canvas) ellipseArc(cx, cy, rx, ry, start, end, width float64, col color.NRGBA, opacity float64) {
	steps := int(math.Ceil((end - start) * math.Max(rx, ry) * 2))
	for i := 0; i < steps; i++ {
		a := start + (end-start)*float64(i)/float64(steps)
		b := start + (end-start)*float64(i+1)/float64(steps)
		c.line(cx+rx*math.Cos(a), cy+ry*math.Sin(a), cx+rx*math.Cos(b), cy+ry*math.Sin(b), width, col, opacity)
	}
}

func bumperEmissionImage() image.Image {
	c := newBumperCanvas()
	// Short independent diffuser segments, dark gaps, narrow hot cores. Cyan is
	// confined to the upper-left sector; most of the assembly retains relay pink.
	for i := 0; i < 20; i++ {
		a := float64(i)*math.Pi/10 + .035
		b := a + math.Pi/10 - .09
		col := pink
		if i >= 11 && i <= 14 {
			col = cyan
		}
		for _, layer := range []struct{ width, opacity float64 }{{10, .035}, {6, .07}, {3.2, .95}, {1.05, .95}} {
			core := col
			if layer.width < 2 {
				core = mix(col, cyanWhite, .8)
			}
			c.ellipseArc(96, 89, 41.8, 37.8, a, b, layer.width, core, layer.opacity)
		}
	}
	return c.finishAlpha()
}

func bumperShadowImage() image.Image {
	c := newBumperCanvas()
	// Cool key from upper-left means the broad penumbra falls down-right.
	c.ellipseSurface(103, 108, 64, 58, func(_, _, r float64) (color.NRGBA, float64) {
		return black, .64 * math.Pow(clamp(1-r*r, 0, 1), .7)
	})
	c.ellipseSurface(97, 99, 58, 57, func(_, _, r float64) (color.NRGBA, float64) {
		return black, .8 * clamp((1-r)*8, 0, 1)
	})
	return c.finishAlpha()
}

func bumperPatchImage() image.Image {
	c := newBumperCanvas()
	// Flat graphite and routed copper around the mounting footprint. A soft
	// rectangular fade avoids a visible decal boundary over the existing board.
	for py := 0; py < c.img.Bounds().Dy(); py++ {
		for px := 0; px < c.img.Bounds().Dx(); px++ {
			x, y := (float64(px)+.5)/c.s-96, (float64(py)+.5)/c.s-96
			fade := clamp((88-math.Max(math.Abs(x), math.Abs(y)))/19, 0, 1)
			grain := math.Sin(x*29+y*71) * math.Sin(x*53-y*23)
			v := 17 + 2.2*grain + 3*bell(x+y, -50, 80)
			c.blendPixel(px, py, materialRGB(v, v*1.22, v*1.37), fade*.92)
		}
	}
	for _, path := range [][][2]float64{
		{{28, 64}, {28, 112}, {56, 140}, {56, 161}, {78, 174}},
		{{19, 59}, {19, 115}, {45, 141}, {45, 162}},
		{{123, 23}, {155, 55}, {155, 123}, {134, 144}, {134, 165}},
		{{116, 18}, {163, 65}, {163, 127}, {148, 142}},
	} {
		for i := 1; i < len(path); i++ {
			a, b := path[i-1], path[i]
			c.line(a[0]+.6, a[1]+.8, b[0]+.6, b[1]+.8, 1.3, black, .7)
			c.line(a[0], a[1], b[0], b[1], .65, metal, .7)
		}
		for _, p := range [][2]float64{path[0], path[len(path)-1]} {
			c.circle(p[0], p[1], 2.4, black, 1)
			c.ring(p[0], p[1], 2.2, 1.5, silver, .65)
		}
	}
	// Local reflected emission is wider and dimmer than the source. Irregular
	// grazing pools read as surface light, not another circular cosmic halo.
	for _, pool := range []struct {
		x, y, rx, ry float64
		col          color.NRGBA
	}{
		{52, 58, 23, 20, cyan}, {147, 114, 20, 31, pink}, {93, 149, 38, 15, pink},
	} {
		c.ellipseSurface(pool.x, pool.y, pool.rx, pool.ry, func(_, _, r float64) (color.NRGBA, float64) {
			falloff := 1 - r
			return pool.col, .17 * falloff * falloff
		})
	}
	return c.finishAlpha()
}

// Average premultiplied samples, then unpremultiply for PNG. Straight-alpha
// averaging would darken pale metal and LED edges against transparent padding.
// Used by every material layer; legacy icons keep their original filter.
func (c *canvas) finishAlpha() *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, c.w, c.h))
	const samples = supersample * supersample
	for y := 0; y < c.h; y++ {
		for x := 0; x < c.w; x++ {
			var r, g, b, a uint64
			for yy := 0; yy < supersample; yy++ {
				for xx := 0; xx < supersample; xx++ {
					i := c.img.PixOffset(x*supersample+xx, y*supersample+yy)
					alpha := uint64(c.img.Pix[i+3])
					r += uint64(c.img.Pix[i]) * alpha
					g += uint64(c.img.Pix[i+1]) * alpha
					b += uint64(c.img.Pix[i+2]) * alpha
					a += alpha
				}
			}
			if a > 0 {
				out.SetNRGBA(x, y, color.NRGBA{R: uint8((r + a/2) / a), G: uint8((g + a/2) / a), B: uint8((b + a/2) / a), A: uint8((a + samples/2) / samples)})
			}
		}
	}
	return out
}
