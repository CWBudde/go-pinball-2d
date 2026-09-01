// Command genassets creates all original image and sound assets used by Neon Relay.
//
// The generator intentionally uses only the Go standard library. All geometry,
// colors, waveforms, and pseudo-random seeds are authored here so that a clean
// checkout can reproduce the committed files byte for byte.
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	supersample = 3
	sampleRate  = 44100
)

var (
	ink       = color.NRGBA{R: 5, G: 8, B: 22, A: 255}
	panel     = color.NRGBA{R: 10, G: 18, B: 43, A: 255}
	cyan      = color.NRGBA{R: 32, G: 232, B: 255, A: 255}
	cyanWhite = color.NRGBA{R: 207, G: 253, B: 255, A: 255}
	magenta   = color.NRGBA{R: 255, G: 43, B: 173, A: 255}
	violet    = color.NRGBA{R: 126, G: 71, B: 255, A: 255}
	steel     = color.NRGBA{R: 111, G: 139, B: 174, A: 255}
)

type canvas struct {
	w, h int
	s    float64
	img  *image.NRGBA
}

func newCanvas(w, h int) *canvas {
	return &canvas{
		w:   w,
		h:   h,
		s:   supersample,
		img: image.NewNRGBA(image.Rect(0, 0, w*supersample, h*supersample)),
	}
}

func (c *canvas) blendPixel(x, y int, src color.NRGBA, opacity float64) {
	if x < 0 || y < 0 || x >= c.img.Bounds().Dx() || y >= c.img.Bounds().Dy() || opacity <= 0 {
		return
	}
	if opacity > 1 {
		opacity = 1
	}
	i := c.img.PixOffset(x, y)
	sa := float64(src.A) / 255 * opacity
	da := float64(c.img.Pix[i+3]) / 255
	oa := sa + da*(1-sa)
	if oa == 0 {
		return
	}
	c.img.Pix[i+0] = uint8(clamp((float64(src.R)*sa+float64(c.img.Pix[i+0])*da*(1-sa))/oa, 0, 255) + 0.5)
	c.img.Pix[i+1] = uint8(clamp((float64(src.G)*sa+float64(c.img.Pix[i+1])*da*(1-sa))/oa, 0, 255) + 0.5)
	c.img.Pix[i+2] = uint8(clamp((float64(src.B)*sa+float64(c.img.Pix[i+2])*da*(1-sa))/oa, 0, 255) + 0.5)
	c.img.Pix[i+3] = uint8(oa*255 + 0.5)
}

func (c *canvas) verticalGradient(top, bottom color.NRGBA) {
	w, h := c.img.Bounds().Dx(), c.img.Bounds().Dy()
	for y := 0; y < h; y++ {
		t := float64(y) / float64(h-1)
		col := mix(top, bottom, t)
		for x := 0; x < w; x++ {
			i := c.img.PixOffset(x, y)
			c.img.Pix[i+0], c.img.Pix[i+1], c.img.Pix[i+2], c.img.Pix[i+3] = col.R, col.G, col.B, col.A
		}
	}
}

func (c *canvas) roundedRect(x0, y0, x1, y1, radius float64, col color.NRGBA, opacity float64) {
	x0, y0, x1, y1, radius = x0*c.s, y0*c.s, x1*c.s, y1*c.s, radius*c.s
	cx, cy := (x0+x1)/2, (y0+y1)/2
	hx, hy := (x1-x0)/2-radius, (y1-y0)/2-radius
	for y := int(math.Floor(y0 - 1)); y <= int(math.Ceil(y1+1)); y++ {
		for x := int(math.Floor(x0 - 1)); x <= int(math.Ceil(x1+1)); x++ {
			qx := math.Abs(float64(x)+0.5-cx) - hx
			qy := math.Abs(float64(y)+0.5-cy) - hy
			d := math.Hypot(math.Max(qx, 0), math.Max(qy, 0)) + math.Min(math.Max(qx, qy), 0) - radius
			coverage := clamp(0.5-d, 0, 1)
			c.blendPixel(x, y, col, opacity*coverage)
		}
	}
}

func (c *canvas) circle(cx, cy, radius float64, col color.NRGBA, opacity float64) {
	c.circleGradient(cx, cy, radius, func(float64) (color.NRGBA, float64) { return col, opacity })
}

func (c *canvas) circleGradient(cx, cy, radius float64, sample func(t float64) (color.NRGBA, float64)) {
	cx, cy, radius = cx*c.s, cy*c.s, radius*c.s
	for y := int(math.Floor(cy - radius - 1)); y <= int(math.Ceil(cy+radius+1)); y++ {
		for x := int(math.Floor(cx - radius - 1)); x <= int(math.Ceil(cx+radius+1)); x++ {
			d := math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy)
			coverage := clamp(radius+0.5-d, 0, 1)
			if coverage > 0 {
				col, opacity := sample(clamp(d/radius, 0, 1))
				c.blendPixel(x, y, col, opacity*coverage)
			}
		}
	}
}

func (c *canvas) ring(cx, cy, outer, inner float64, col color.NRGBA, opacity float64) {
	cx, cy, outer, inner = cx*c.s, cy*c.s, outer*c.s, inner*c.s
	for y := int(math.Floor(cy - outer - 1)); y <= int(math.Ceil(cy+outer+1)); y++ {
		for x := int(math.Floor(cx - outer - 1)); x <= int(math.Ceil(cx+outer+1)); x++ {
			d := math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy)
			coverage := clamp(outer+0.5-d, 0, 1) * clamp(d-inner+0.5, 0, 1)
			c.blendPixel(x, y, col, opacity*coverage)
		}
	}
}

func (c *canvas) line(x0, y0, x1, y1, width float64, col color.NRGBA, opacity float64) {
	x0, y0, x1, y1, width = x0*c.s, y0*c.s, x1*c.s, y1*c.s, width*c.s
	r := width / 2
	minX, maxX := math.Min(x0, x1)-r-1, math.Max(x0, x1)+r+1
	minY, maxY := math.Min(y0, y1)-r-1, math.Max(y0, y1)+r+1
	for y := int(math.Floor(minY)); y <= int(math.Ceil(maxY)); y++ {
		for x := int(math.Floor(minX)); x <= int(math.Ceil(maxX)); x++ {
			d := pointSegmentDistance(float64(x)+0.5, float64(y)+0.5, x0, y0, x1, y1)
			c.blendPixel(x, y, col, opacity*clamp(r+0.5-d, 0, 1))
		}
	}
}

func (c *canvas) polygon(points [][2]float64, col color.NRGBA, opacity float64) {
	if len(points) < 3 {
		return
	}
	minX, maxX, minY, maxY := points[0][0], points[0][0], points[0][1], points[0][1]
	p := make([][2]float64, len(points))
	for i, point := range points {
		p[i] = [2]float64{point[0] * c.s, point[1] * c.s}
		minX, maxX = math.Min(minX, point[0]), math.Max(maxX, point[0])
		minY, maxY = math.Min(minY, point[1]), math.Max(maxY, point[1])
	}
	for y := int(math.Floor(minY * c.s)); y <= int(math.Ceil(maxY*c.s)); y++ {
		for x := int(math.Floor(minX * c.s)); x <= int(math.Ceil(maxX*c.s)); x++ {
			if insidePolygon(float64(x)+0.5, float64(y)+0.5, p) {
				c.blendPixel(x, y, col, opacity)
			}
		}
	}
}

func (c *canvas) glowLine(x0, y0, x1, y1, width float64, col color.NRGBA, opacity float64) {
	c.line(x0, y0, x1, y1, width*6, col, opacity*0.035)
	c.line(x0, y0, x1, y1, width*3, col, opacity*0.11)
	c.line(x0, y0, x1, y1, width, col, opacity)
}

func (c *canvas) finish() *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, c.w, c.h))
	ss := supersample * supersample
	for y := 0; y < c.h; y++ {
		for x := 0; x < c.w; x++ {
			var r, g, b, a int
			for yy := 0; yy < supersample; yy++ {
				for xx := 0; xx < supersample; xx++ {
					i := c.img.PixOffset(x*supersample+xx, y*supersample+yy)
					r += int(c.img.Pix[i+0])
					g += int(c.img.Pix[i+1])
					b += int(c.img.Pix[i+2])
					a += int(c.img.Pix[i+3])
				}
			}
			i := out.PixOffset(x, y)
			out.Pix[i+0], out.Pix[i+1], out.Pix[i+2], out.Pix[i+3] = uint8(r/ss), uint8(g/ss), uint8(b/ss), uint8(a/ss)
		}
	}
	return out
}

type rng uint64

func (r *rng) next() uint64 {
	x := uint64(*r)
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	*r = rng(x)
	return x
}

func (r *rng) unit() float64 { return float64(r.next()>>11) / float64(uint64(1)<<53) }

func backgroundImage() image.Image {
	c := newCanvas(720, 1080)
	c.verticalGradient(color.NRGBA{R: 3, G: 7, B: 21, A: 255}, color.NRGBA{R: 12, G: 8, B: 39, A: 255})

	// Recessed playfield panels and a restrained technical grid.
	c.roundedRect(25, 25, 695, 1055, 42, color.NRGBA{R: 10, G: 20, B: 47, A: 255}, 0.96)
	c.roundedRect(37, 37, 683, 1043, 34, color.NRGBA{R: 3, G: 10, B: 29, A: 255}, 0.78)
	for x := 48.0; x < 680; x += 48 {
		c.line(x, 45, x, 1037, 1, cyan, 0.045)
	}
	for y := 55.0; y < 1030; y += 48 {
		c.line(43, y, 677, y, 1, violet, 0.045)
	}

	// Hand-authored circuit routes, mirrored to frame the open table center.
	routes := [][][2]float64{
		{{58, 150}, {112, 150}, {138, 124}, {205, 124}},
		{{58, 222}, {94, 222}, {123, 251}, {185, 251}, {211, 277}},
		{{58, 340}, {112, 340}, {139, 313}, {190, 313}},
		{{58, 486}, {99, 486}, {128, 515}, {198, 515}},
		{{58, 624}, {115, 624}, {143, 596}, {208, 596}},
		{{58, 760}, {89, 760}, {125, 796}, {196, 796}},
		{{58, 908}, {112, 908}, {142, 878}, {210, 878}},
	}
	for i, route := range routes {
		var col color.NRGBA
		switch i % 3 {
		case 0:
			col = cyan
		case 1:
			col = magenta
		case 2:
			col = violet
		}
		drawRoute(c, route, col, 0.44)
		mirror := make([][2]float64, len(route))
		for j, p := range route {
			mirror[j] = [2]float64{720 - p[0], p[1] + float64((i%2)*13-6)}
		}
		drawRoute(c, mirror, col, 0.38)
	}

	seed := rng(0x4e454f4e52454c59) // "NEONRELY"
	for i := 0; i < 90; i++ {
		x := 55 + seed.unit()*610
		y := 55 + seed.unit()*970
		r := 1.2 + seed.unit()*1.8
		col := cyan
		if i%4 == 0 {
			col = magenta
		}
		c.circle(x, y, r*3.2, col, 0.04)
		c.circle(x, y, r, col, 0.32)
	}

	// Central relay motif and launch-lane energy rails.
	for i := 0; i < 5; i++ {
		r := 78 + float64(i)*22
		c.ring(360, 355, r+1.2, r-1.2, violet, 0.08+float64(i)*0.012)
	}
	c.glowLine(631, 150, 631, 928, 2, cyan, 0.48)
	c.glowLine(656, 178, 656, 946, 1.4, magenta, 0.36)
	for y := 190.0; y < 930; y += 72 {
		c.line(620, y, 642, y, 2, cyanWhite, 0.42)
	}

	// Edge illumination and a subtle lower drain chevron.
	c.ring(360, 540, 660, 654, cyan, 0.11)
	c.line(268, 998, 360, 1040, 3, magenta, 0.20)
	c.line(360, 1040, 452, 998, 3, magenta, 0.20)
	return c.finish()
}

func drawRoute(c *canvas, route [][2]float64, col color.NRGBA, opacity float64) {
	for i := 1; i < len(route); i++ {
		c.glowLine(route[i-1][0], route[i-1][1], route[i][0], route[i][1], 1.5, col, opacity)
	}
	for _, p := range route {
		c.circle(p[0], p[1], 4.5, col, opacity*0.23)
		c.ring(p[0], p[1], 2.6, 1.25, col, opacity)
	}
}

var glyphs = map[rune][7]string{
	'N': {"10001", "11001", "11001", "10101", "10011", "10011", "10001"},
	'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
	'O': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
	'R': {"11110", "10001", "10001", "11110", "10100", "10010", "10001"},
	'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
	'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
	'Y': {"10001", "10001", "01010", "00100", "00100", "00100", "00100"},
}

func bitmapText(c *canvas, text string, x, y, unit float64, col color.NRGBA) {
	for _, ch := range text {
		glyph := glyphs[ch]
		for row, line := range glyph {
			for column, pixel := range line {
				if pixel != '1' {
					continue
				}
				x0 := x + float64(column)*unit
				y0 := y + float64(row)*unit
				c.roundedRect(x0-unit*.12, y0-unit*.12, x0+unit*.82, y0+unit*.82, unit*.34, col, 0.055)
				c.roundedRect(x0+unit*.08, y0+unit*.08, x0+unit*.62, y0+unit*.62, unit*.18, col, 0.18)
				c.roundedRect(x0+unit*.24, y0+unit*.24, x0+unit*.46, y0+unit*.46, unit*.08, cyanWhite, 0.96)
			}
		}
		x += 6 * unit
	}
}

func logoImage() image.Image {
	c := newCanvas(640, 200)
	// Relay traces behind the wordmark.
	c.glowLine(35, 37, 118, 37, 2.4, magenta, 0.6)
	c.glowLine(522, 160, 605, 160, 2.4, cyan, 0.6)
	c.line(35, 37, 35, 75, 2.4, magenta, 0.65)
	c.line(605, 124, 605, 160, 2.4, cyan, 0.65)
	c.ring(35, 80, 6, 3, magenta, 0.8)
	c.ring(605, 119, 6, 3, cyan, 0.8)
	bitmapText(c, "NEON", 113, 18, 18, cyan)
	bitmapText(c, "RELAY", 140, 112, 12, magenta)
	c.glowLine(218, 99, 422, 99, 2, violet, 0.6)
	for _, x := range []float64{218, 320, 422} {
		c.circle(x, 99, 4.5, violet, 0.7)
	}
	return c.finish()
}

func faviconImage() image.Image {
	c := newCanvas(64, 64)
	c.circleGradient(32, 32, 30, func(t float64) (color.NRGBA, float64) {
		return mix(color.NRGBA{R: 23, G: 31, B: 70, A: 255}, ink, t), 1
	})
	c.ring(32, 32, 29, 26.5, cyan, 0.75)
	c.glowLine(18, 46, 18, 18, 4, cyan, 0.9)
	c.glowLine(18, 18, 46, 46, 4, magenta, 0.9)
	c.glowLine(46, 46, 46, 18, 4, cyan, 0.9)
	c.circle(18, 18, 3, cyanWhite, 0.95)
	c.circle(46, 46, 3, cyanWhite, 0.95)
	return c.finish()
}

func ballImage() image.Image {
	c := newCanvas(64, 64)
	c.circle(32, 34, 28, cyan, 0.05)
	c.circleGradient(32, 31, 25.5, func(t float64) (color.NRGBA, float64) {
		if t < .63 {
			return mix(color.NRGBA{R: 240, G: 255, B: 255, A: 255}, steel, t/.63), 1
		}
		return mix(steel, color.NRGBA{R: 22, G: 39, B: 67, A: 255}, (t-.63)/.37), 1
	})
	c.circle(23, 21, 7.5, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, 0.72)
	c.circle(20, 18, 2.8, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, 0.95)
	c.ring(32, 31, 25.7, 24.6, cyanWhite, 0.4)
	return c.finish()
}

func flipperImage() image.Image {
	c := newCanvas(180, 64)
	c.line(30, 32, 148, 32, 42, magenta, 0.07)
	c.line(30, 32, 148, 32, 34, color.NRGBA{R: 36, G: 14, B: 61, A: 255}, 1)
	c.line(30, 32, 148, 32, 28, magenta, 0.92)
	c.line(30, 32, 148, 32, 18, color.NRGBA{R: 255, G: 113, B: 211, A: 255}, 0.9)
	c.line(32, 26, 148, 26, 4, color.NRGBA{R: 255, G: 229, B: 249, A: 255}, 0.55)
	c.circle(30, 32, 19, panel, 1)
	c.ring(30, 32, 18, 13, cyan, 0.9)
	c.circle(30, 32, 8, cyanWhite, 0.9)
	return c.finish()
}

func bumperImage() image.Image {
	c := newCanvas(128, 128)
	c.circleGradient(64, 66, 57, func(t float64) (color.NRGBA, float64) { return violet, .10 * (1 - t) })
	c.circle(64, 64, 48, color.NRGBA{R: 14, G: 21, B: 50, A: 255}, 1)
	c.ring(64, 64, 48, 42, magenta, 0.92)
	c.circleGradient(64, 64, 38, func(t float64) (color.NRGBA, float64) {
		return mix(color.NRGBA{R: 64, G: 39, B: 109, A: 255}, color.NRGBA{R: 14, G: 12, B: 39, A: 255}, t), 1
	})
	c.ring(64, 64, 31, 27, cyan, 0.88)
	c.circle(64, 64, 18, magenta, 0.18)
	c.polygon([][2]float64{{64, 43}, {70, 57}, {84, 64}, {70, 71}, {64, 85}, {58, 71}, {44, 64}, {58, 57}}, cyanWhite, 0.9)
	c.circle(52, 47, 6, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, 0.35)
	return c.finish()
}

func postImage() image.Image {
	c := newCanvas(48, 48)
	c.circle(24, 25, 21, cyan, 0.08)
	c.circleGradient(24, 24, 17, func(t float64) (color.NRGBA, float64) {
		return mix(cyanWhite, color.NRGBA{R: 32, G: 81, B: 113, A: 255}, t), 1
	})
	c.ring(24, 24, 17.5, 14.5, cyan, 0.92)
	c.circle(20, 19, 4, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, 0.55)
	return c.finish()
}

func targetImage() image.Image {
	c := newCanvas(64, 96)
	c.roundedRect(8, 5, 56, 91, 8, magenta, 0.12)
	c.roundedRect(11, 7, 53, 89, 7, color.NRGBA{R: 30, G: 17, B: 55, A: 255}, 1)
	c.roundedRect(14, 10, 50, 86, 5, magenta, 0.88)
	c.roundedRect(19, 15, 45, 81, 3, color.NRGBA{R: 73, G: 18, B: 72, A: 255}, 1)
	c.polygon([][2]float64{{32, 25}, {39, 40}, {47, 48}, {39, 56}, {32, 71}, {25, 56}, {17, 48}, {25, 40}}, cyanWhite, 0.92)
	c.line(19, 18, 45, 18, 2, color.NRGBA{R: 255, G: 228, B: 247, A: 255}, 0.6)
	return c.finish()
}

func laneLightImage() image.Image {
	c := newCanvas(48, 96)
	c.line(24, 17, 24, 78, 14, cyan, 0.055)
	c.roundedRect(16, 8, 32, 88, 8, color.NRGBA{R: 8, G: 29, B: 47, A: 255}, 0.9)
	c.roundedRect(19, 11, 29, 85, 5, cyan, 0.48)
	for _, y := range []float64{22, 40, 58, 76} {
		c.circle(24, y, 8, cyan, 0.12)
		c.polygon([][2]float64{{24, y - 5}, {29, y + 2}, {26, y + 2}, {26, y + 6}, {22, y + 6}, {22, y + 2}, {19, y + 2}}, cyanWhite, 0.9)
	}
	return c.finish()
}

func plungerImage() image.Image {
	c := newCanvas(56, 180)
	c.roundedRect(12, 9, 44, 170, 14, cyan, 0.06)
	c.roundedRect(18, 8, 38, 132, 9, color.NRGBA{R: 18, G: 34, B: 60, A: 255}, 1)
	c.roundedRect(21, 11, 35, 130, 6, steel, 0.9)
	c.line(22, 27, 34, 27, 3, cyanWhite, 0.62)
	c.line(22, 44, 34, 44, 3, cyanWhite, 0.52)
	c.line(22, 61, 34, 61, 3, cyanWhite, 0.44)
	c.line(22, 78, 34, 78, 3, cyanWhite, 0.36)
	c.line(28, 128, 28, 151, 8, magenta, 0.88)
	c.roundedRect(8, 145, 48, 175, 12, color.NRGBA{R: 52, G: 17, B: 66, A: 255}, 1)
	c.roundedRect(11, 148, 45, 172, 10, magenta, 0.9)
	c.line(16, 153, 40, 153, 3, color.NRGBA{R: 255, G: 220, B: 245, A: 255}, 0.55)
	return c.finish()
}

func glowImage() image.Image {
	c := newCanvas(192, 192)
	c.circleGradient(96, 96, 94, func(t float64) (color.NRGBA, float64) {
		a := (1 - t) * (1 - t) * (1 - t) * 0.55
		return mix(cyan, violet, t), a
	})
	c.circleGradient(96, 96, 45, func(t float64) (color.NRGBA, float64) {
		return mix(cyanWhite, cyan, t), (1 - t) * (1 - t) * .34
	})
	return c.finish()
}

func particleImage() image.Image {
	c := newCanvas(32, 32)
	c.circleGradient(16, 16, 15, func(t float64) (color.NRGBA, float64) {
		return cyan, (1 - t) * (1 - t) * .18
	})
	c.polygon([][2]float64{{16, 1}, {19, 12}, {31, 16}, {19, 20}, {16, 31}, {13, 20}, {1, 16}, {13, 12}}, cyan, 0.78)
	c.polygon([][2]float64{{16, 7}, {18, 14}, {25, 16}, {18, 18}, {16, 25}, {14, 18}, {7, 16}, {14, 14}}, cyanWhite, 0.95)
	c.circle(16, 16, 3, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, 1)
	return c.finish()
}

type audioBuilder struct {
	samples []float64
	seed    rng
}

func newAudio(seconds float64, seed uint64) *audioBuilder {
	return &audioBuilder{samples: make([]float64, int(seconds*sampleRate)), seed: rng(seed)}
}

func (a *audioBuilder) tone(start, duration, frequency, endFrequency, amplitude float64, wave func(float64) float64) {
	from := int(start * sampleRate)
	to := min(len(a.samples), int((start+duration)*sampleRate))
	phase := 0.0
	for i := from; i < to; i++ {
		t := float64(i-from) / float64(sampleRate)
		u := clamp(t/duration, 0, 1)
		freq := frequency + (endFrequency-frequency)*u
		phase += 2 * math.Pi * freq / sampleRate
		env := (1 - u) * (1 - u) * smoothAttack(u, .025)
		a.samples[i] += wave(phase) * amplitude * env
	}
}

func (a *audioBuilder) noise(start, duration, amplitude float64, decay float64) {
	from := int(start * sampleRate)
	to := min(len(a.samples), int((start+duration)*sampleRate))
	state := 0.0
	for i := from; i < to; i++ {
		u := float64(i-from) / float64(max(1, to-from-1))
		white := a.seed.unit()*2 - 1
		state = state*.54 + white*.46
		env := math.Pow(1-u, decay) * smoothAttack(u, .008)
		a.samples[i] += state * amplitude * env
	}
}

func (a *audioBuilder) click(start, amplitude float64) {
	from := int(start * sampleRate)
	for j := 0; j < 120 && from+j < len(a.samples); j++ {
		u := float64(j) / 120
		a.samples[from+j] += (a.seed.unit()*2 - 1) * amplitude * (1 - u) * (1 - u) * (1 - u)
	}
}

func (a *audioBuilder) wav() []byte {
	// Remove DC, apply a short master fade, and normalize without changing shape.
	mean := 0.0
	for _, v := range a.samples {
		mean += v
	}
	mean /= float64(len(a.samples))
	peak := 0.0
	for i := range a.samples {
		a.samples[i] -= mean
		fade := min(1, float64(len(a.samples)-1-i)/(sampleRate*.012))
		a.samples[i] *= clamp(fade, 0, 1)
		peak = math.Max(peak, math.Abs(a.samples[i]))
	}
	scale := .91
	if peak > 0 {
		scale /= peak
	}
	var out bytes.Buffer
	dataSize := len(a.samples) * 2
	out.WriteString("RIFF")
	_ = binary.Write(&out, binary.LittleEndian, uint32(36+dataSize))
	out.WriteString("WAVEfmt ")
	_ = binary.Write(&out, binary.LittleEndian, uint32(16))
	_ = binary.Write(&out, binary.LittleEndian, uint16(1))
	_ = binary.Write(&out, binary.LittleEndian, uint16(1))
	_ = binary.Write(&out, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(&out, binary.LittleEndian, uint32(sampleRate*2))
	_ = binary.Write(&out, binary.LittleEndian, uint16(2))
	_ = binary.Write(&out, binary.LittleEndian, uint16(16))
	out.WriteString("data")
	_ = binary.Write(&out, binary.LittleEndian, uint32(dataSize))
	for _, v := range a.samples {
		pcm := int16(clamp(v*scale, -1, 1) * 32767)
		_ = binary.Write(&out, binary.LittleEndian, pcm)
	}
	return out.Bytes()
}

func sine(p float64) float64 { return math.Sin(p) }
func triangle(p float64) float64 {
	return 2 / math.Pi * math.Asin(math.Sin(p))
}

func flipperSound() []byte {
	a := newAudio(.16, 0xf11f)
	a.click(0, .8)
	a.noise(0, .08, .42, 3.2)
	a.tone(0, .14, 112, 67, .7, sine)
	a.tone(.012, .10, 310, 190, .24, triangle)
	return a.wav()
}

func bumperSound() []byte {
	a := newAudio(.31, 0xb00b)
	a.click(0, .42)
	a.tone(0, .29, 630, 570, .75, sine)
	a.tone(.008, .26, 945, 855, .40, sine)
	a.tone(.016, .22, 1260, 1140, .19, sine)
	return a.wav()
}

func targetSound() []byte {
	a := newAudio(.20, 0x7a267e7)
	a.click(0, .72)
	a.noise(0, .09, .34, 2.7)
	a.tone(0, .16, 390, 235, .48, triangle)
	return a.wav()
}

func launchSound() []byte {
	a := newAudio(.62, 0x1a0c4)
	a.noise(0, .28, .25, .45)
	a.tone(0, .46, 105, 510, .48, sine)
	a.tone(.06, .38, 190, 890, .22, triangle)
	a.click(.37, .7)
	a.noise(.37, .22, .4, 2.4)
	return a.wav()
}

func jackpotSound() []byte {
	a := newAudio(1.34, 0xacce55)
	notes := []float64{523.25, 659.25, 783.99, 1046.50, 1318.51}
	for i, note := range notes {
		start := float64(i) * .145
		a.tone(start, .52, note, note*1.012, .54, sine)
		a.tone(start, .45, note*2, note*2.006, .16, sine)
		a.click(start, .12)
	}
	for _, note := range []float64{523.25, 659.25, 783.99} {
		a.tone(.75, .56, note, note, .22, sine)
	}
	return a.wav()
}

func drainSound() []byte {
	a := newAudio(.92, 0xd2a10)
	a.noise(0, .12, .28, 2)
	a.tone(0, .88, 330, 72, .64, sine)
	a.tone(.07, .72, 247, 55, .38, triangle)
	return a.wav()
}

func gameOverSound() []byte {
	a := newAudio(1.65, 0x6a6e0f)
	starts := []float64{0, .38, .76, 1.12}
	roots := []float64{392, 349.23, 293.66, 220}
	for i, root := range roots {
		a.tone(starts[i], .49, root, root*.96, .42, sine)
		a.tone(starts[i], .49, root*1.1892, root*1.142, .28, sine)
		a.tone(starts[i], .49, root*1.4983, root*1.438, .22, sine)
	}
	a.noise(1.12, .31, .08, 2.4)
	return a.wav()
}

type generatedFile struct {
	path string
	data []byte
}

func generatedAssets() ([]generatedFile, error) {
	images := []struct {
		path string
		make func() image.Image
	}{
		{"images/background.png", backgroundImage},
		{"images/logo.png", logoImage},
		{"images/favicon.png", faviconImage},
		{"images/ball.png", ballImage},
		{"images/flipper.png", flipperImage},
		{"images/bumper.png", bumperImage},
		{"images/post.png", postImage},
		{"images/target.png", targetImage},
		{"images/lane-light.png", laneLightImage},
		{"images/plunger.png", plungerImage},
		{"images/glow.png", glowImage},
		{"images/particle.png", particleImage},
	}
	files := make([]generatedFile, 0, len(images)+7)
	for _, item := range images {
		var buf bytes.Buffer
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(&buf, item.make()); err != nil {
			return nil, fmt.Errorf("encode %s: %w", item.path, err)
		}
		files = append(files, generatedFile{item.path, buf.Bytes()})
	}
	audio := []struct {
		path string
		make func() []byte
	}{
		{"audio/flipper.wav", flipperSound},
		{"audio/bumper.wav", bumperSound},
		{"audio/target.wav", targetSound},
		{"audio/launch.wav", launchSound},
		{"audio/jackpot.wav", jackpotSound},
		{"audio/drain.wav", drainSound},
		{"audio/game-over.wav", gameOverSound},
	}
	for _, item := range audio {
		files = append(files, generatedFile{item.path, item.make()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].path < files[j].path })
	return files, nil
}

func writeAssets(root string, files []generatedFile) error {
	for _, file := range files {
		path := filepath.Join(root, filepath.FromSlash(file.path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, file.data, 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote %s (%d bytes)\n", path, len(file.data))
	}
	return nil
}

func checkAssets(root string, files []generatedFile) error {
	var problems []string
	for _, expected := range files {
		path := filepath.Join(root, filepath.FromSlash(expected.path))
		actual, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		if !bytes.Equal(actual, expected.data) {
			problems = append(problems, path+": stale (run go run ./cmd/genassets)")
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("asset check failed:\n  %s", strings.Join(problems, "\n  "))
	}
	return nil
}

func main() {
	out := flag.String("out", "assets", "asset output directory")
	check := flag.Bool("check", false, "verify committed assets without rewriting them")
	flag.Parse()
	files, err := generatedAssets()
	if err == nil {
		if *check {
			err = checkAssets(*out, files)
		} else {
			err = writeAssets(*out, files)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *check {
		fmt.Printf("verified %d generated assets in %s\n", len(files), *out)
	}
}

func pointSegmentDistance(px, py, ax, ay, bx, by float64) float64 {
	dx, dy := bx-ax, by-ay
	denom := dx*dx + dy*dy
	if denom == 0 {
		return math.Hypot(px-ax, py-ay)
	}
	t := clamp(((px-ax)*dx+(py-ay)*dy)/denom, 0, 1)
	return math.Hypot(px-(ax+t*dx), py-(ay+t*dy))
}

func insidePolygon(x, y float64, points [][2]float64) bool {
	inside := false
	j := len(points) - 1
	for i := range points {
		xi, yi := points[i][0], points[i][1]
		xj, yj := points[j][0], points[j][1]
		if (yi > y) != (yj > y) && x < (xj-xi)*(y-yi)/(yj-yi)+xi {
			inside = !inside
		}
		j = i
	}
	return inside
}

func mix(a, b color.NRGBA, t float64) color.NRGBA {
	t = clamp(t, 0, 1)
	lerp := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + .5) }
	return color.NRGBA{R: lerp(a.R, b.R), G: lerp(a.G, b.G), B: lerp(a.B, b.B), A: lerp(a.A, b.A)}
}

func smoothAttack(t, length float64) float64 {
	if t >= length {
		return 1
	}
	u := clamp(t/length, 0, 1)
	return u * u * (3 - 2*u)
}

func clamp(v, low, high float64) float64 { return math.Max(low, math.Min(high, v)) }
