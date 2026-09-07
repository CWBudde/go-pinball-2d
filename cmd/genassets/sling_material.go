package main

import (
	"image/color"
	"math"

	"github.com/CWBudde/go-pinball-2d/internal/table"
)

// Slings share their outline with the collider. Raised material stays inside
// the nine-unit rubber radius; only the existing shadow extends into the lanes.
func (c *canvas) plastic(p [][2]float64) {
	cx, cy := (p[0][0]+p[1][0]+p[2][0])/3, (p[0][1]+p[1][1]+p[2][1])/3
	inset := func(scale, dy float64) [][2]float64 {
		q := make([][2]float64, 3)
		for i, v := range p {
			q[i] = [2]float64{cx + (v[0]-cx)*scale, cy + (v[1]-cy)*scale + dy}
		}
		return q
	}
	accent := cyan
	if cx > table.PlayfieldCenter {
		accent = pink
	}
	// A molded black skirt supports a recessed red diffuser and a separate
	// polished cover. All offsets plus half-widths remain below nine units.
	c.polygon(p, black, 1)
	c.outline(p, 18, black, 1)
	c.outline(p, 16, materialRGB(38, 20, 27), 1)
	c.outline(p, 13, materialRGB(12, 13, 16), 1)
	a, b := p[1], p[2]
	c.line(a[0], a[1]+3, b[0], b[1]+3, 8, materialRGB(83, 13, 35), 1)
	for i := 0; i < 5; i++ {
		u, v := .12+float64(i)*.155, .235+float64(i)*.155
		x0, y0 := a[0]+(b[0]-a[0])*u, a[1]+(b[1]-a[1])*u+5
		x1, y1 := a[0]+(b[0]-a[0])*v, a[1]+(b[1]-a[1])*v+5
		c.glowLine(x0, y0, x1, y1, 2, pink, .8)
		c.line(x0, y0-.4, x1, y1-.4, .6, cyanWhite, .85)
	}
	// The upper-left key illuminates each edge according to its orientation,
	// rather than giving all three sides the same bright outline.
	for i, a := range p {
		b := p[(i+1)%3]
		c.capsuleSurface(a[0], a[1]-1.5, b[0], b[1]-1.5, 5.2, bumperChrome)
		c.line(a[0], a[1]-2.5, b[0], b[1]-2.5, 1, silver, .75)
	}
	q := inset(.93, -1.5)
	c.polygon(q, black, 1)
	c.outline(q, 2.4, black, 1)
	c.slingGlass(q, accent)
	c.outline(q, .7, metal, .75)
	// Bent traces, plated vias, and miniature solder pads sit beneath the
	// smoked acrylic. Barycentric coordinates keep every detail within the cover.
	point := func(a, b float64) [2]float64 {
		return [2]float64{q[0][0]*a + q[1][0]*b + q[2][0]*(1-a-b), q[0][1]*a + q[1][1]*b + q[2][1]*(1-a-b)}
	}
	for _, weights := range [][4][2]float64{
		{{.76, .12}, {.49, .14}, {.33, .31}, {.12, .72}},
		{{.55, .11}, {.29, .14}, {.18, .31}, {.13, .37}},
	} {
		for i := 0; i < len(weights)-1; i++ {
			a, b := point(weights[i][0], weights[i][1]), point(weights[i+1][0], weights[i+1][1])
			c.glowLine(a[0], a[1], b[0], b[1], .9, accent, .65)
			c.line(a[0]-.4, a[1]-.4, b[0]-.4, b[1]-.4, .35, silver, .35)
		}
		for _, w := range [][2]float64{weights[0], weights[3]} {
			v := point(w[0], w[1])
			c.circle(v[0], v[1], 2.3, black, 1)
			c.ring(v[0], v[1], 1.8, 1.1, accent, .9)
			c.circle(v[0]-.5, v[1]-.5, .5, silver, .8)
		}
	}
	for _, v := range p {
		c.ellipseSurface(v[0], v[1]+1, 7.8, 7.8, rubberMaterial)
		c.ellipseSurface(v[0], v[1]-1, 6.7, 6.7, bumperChrome)
		c.bumperFastener(v[0], v[1]-1)
	}
}

// Shade only the convex cover face. Subpixel sampling is provided by the
// material canvas; grain and fine scratches are deterministic, like the bumpers.
func (c *canvas) slingGlass(p [][2]float64, accent color.NRGBA) {
	minX, maxX, minY, maxY := p[0][0], p[0][0], p[0][1], p[0][1]
	for _, v := range p[1:] {
		minX, maxX = math.Min(minX, v[0]), math.Max(maxX, v[0])
		minY, maxY = math.Min(minY, v[1]), math.Max(maxY, v[1])
	}
	for py := int(minY * c.s); py <= int(maxY*c.s); py++ {
		for px := int(minX * c.s); px <= int(maxX*c.s); px++ {
			x, y := (float64(px)+.5)/c.s, (float64(py)+.5)/c.s
			positive, negative, distance := false, false, math.Inf(1)
			for i, a := range p {
				b := p[(i+1)%3]
				cross := (b[0]-a[0])*(y-a[1]) - (b[1]-a[1])*(x-a[0])
				positive = positive || cross > 0
				negative = negative || cross < 0
				distance = math.Min(distance, pointSegmentDistance(x, y, a[0], a[1], b[0], b[1]))
			}
			if positive && negative {
				continue
			}
			v := 13 + 10*(maxY-y)/(maxY-minY)
			v += 18*bell(y+.7*x, minY+.7*minX+50, 8) + 2*math.Sin(x*31+y*17)*math.Sin(y*53)
			v += 7 * bell(distance, 1.1, .5)
			reflection := 9 * math.Exp(-distance/4)
			col := materialRGB(v+reflection*float64(accent.R)/255, v*1.08+reflection*float64(accent.G)/255, v*1.2+reflection*float64(accent.B)/255)
			c.blendPixel(px, py, col, 1)
		}
	}
}
