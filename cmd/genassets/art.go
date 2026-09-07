package main

import (
	"image"
	"image/color"
	"math"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
)

// The static hardware is drawn from the same definition as the physics. These
// rails and plastics are baked once, avoiding hundreds of canvas calls per frame.
var (
	graphite = color.NRGBA{R: 22, G: 28, B: 34, A: 255}
	metal    = color.NRGBA{R: 97, G: 113, B: 123, A: 255}
	silver   = color.NRGBA{R: 201, G: 215, B: 220, A: 255}
	black    = color.NRGBA{R: 3, G: 6, B: 9, A: 255}
	pink     = color.NRGBA{R: 255, G: 48, B: 129, A: 255}
	gold     = color.NRGBA{R: 255, G: 184, B: 66, A: 255}
)

func backgroundImage() image.Image {
	c := newCanvas(720, 1080)
	c.verticalGradient(graphite, black)

	// Subtle etched grain, with no free-floating stars or space artwork.
	seed := rng(0x52454c4159)
	for y := 42.0; y < 1040; y += 3 {
		for x := 43.0; x < 680; x += 4 {
			c.line(x, y, x+2, y, .5, silver, .015+seed.unit()*.022)
		}
	}
	// Broad dark routing panels frame an open central field.
	for _, p := range [][][2]float64{
		{{65, 185}, {155, 95}, {280, 95}, {280, 190}, {170, 300}, {170, 600}, {110, 660}, {65, 660}},
		{{550, 200}, {600, 200}, {600, 675}, {550, 675}, {550, 620}, {575, 590}, {575, 370}, {550, 345}},
		{{224, 423}, {417, 423}, {442, 448}, {442, 524}, {416, 551}, {224, 551}, {205, 528}, {205, 448}},
	} {
		c.polygon(p, black, .35)
		c.outline(p, 1, metal, .28)
	}
	routes := [][][2]float64{
		{{90, 180}, {150, 120}, {175, 120}},
		{{85, 260}, {115, 230}, {115, 195}, {157, 153}},
		{{74, 390}, {106, 358}, {106, 307}, {169, 244}, {169, 216}},
		{{76, 480}, {128, 428}, {128, 331}, {185, 274}},
		{{80, 598}, {147, 531}, {147, 404}, {190, 361}},
		{{113, 640}, {180, 573}, {180, 481}, {239, 422}, {239, 380}},
		{{160, 670}, {202, 628}, {202, 531}, {267, 466}},
		{{286, 174}, {286, 206}, {322, 242}, {322, 351}, {299, 374}},
	}
	for i, route := range routes {
		col := cyan
		if i%3 == 1 {
			col = pink
		}
		drawRoute(c, route, col, .32)
		parallel := make([][2]float64, len(route))
		mirror := make([][2]float64, len(route))
		for j, p := range route {
			parallel[j] = [2]float64{p[0] + 8, p[1] + 8}
			mirror[j] = [2]float64{650 - p[0], p[1]}
		}
		drawRoute(c, parallel, metal, .25)
		drawRoute(c, mirror, col, .24)
	}
	// Screen-printed center identity and a real switching-contact diagram.
	c.wordmark("NEON", table.TitleOrigin.X, table.TitleOrigin.Y, 8.0, cyan)
	c.wordmark("RELAY", table.TitleOrigin.X-4, table.TitleOrigin.Y+55, 6.4, pink)
	c.relay(344, table.TitleOrigin.Y+116, .9, cyan)
	for i := 0; i < 3; i++ {
		y := 764.0 + float64(i)*23
		c.glowLine(341, y+11, 360, y, 2, pink, .55)
		c.glowLine(360, y, 379, y+11, 2, pink, .55)
	}

	for i := 0; i < 9; i++ {
		y := 305.0 + float64(i)*69
		c.circle(643, y, 4, black, 1)
		c.ring(643, y, 3, 1.7, gold, .7)
		if i < 8 {
			c.line(643, y+6, 643, y+62, 1, metal, .45)
		}
	}
	return c.finish()
}

func (c *canvas) outline(p [][2]float64, width float64, col color.NRGBA, opacity float64) {
	for i, a := range p {
		b := p[(i+1)%len(p)]
		c.line(a[0], a[1], b[0], b[1], width, col, opacity)
	}
}

func (c *canvas) rails(walls []physics.LineCollider) {
	// Draw each material across the whole run before the next layer, so joins
	// read as continuous rails rather than individually capped line segments.
	for _, layer := range []struct {
		dx, dy, extra, opacity float64
		col                    color.NRGBA
	}{
		{2, 5, 9, .6, black},
		{0, 0, 5, 1, black},
		{0, 0, 0, 1, metal},
		{-1, -1, -3, .9, silver},
		{0, 0, -5, 1, graphite},
	} {
		for _, wall := range walls {
			a, b := wall.Segment.A, wall.Segment.B
			c.line(a.X+layer.dx, a.Y+layer.dy, b.X+layer.dx, b.Y+layer.dy, math.Max(2, wall.Radius*2+layer.extra), layer.col, layer.opacity)
		}
	}
	for _, wall := range walls {
		a, b := wall.Segment.A, wall.Segment.B
		c.glowLine(a.X, a.Y, b.X, b.Y, 1.2, cyan, .6)
	}
}

func (c *canvas) plastic(p [][2]float64) {
	shadow := make([][2]float64, len(p))
	inner := make([][2]float64, len(p))
	var cx, cy float64
	for _, v := range p {
		cx += v[0] / float64(len(p))
		cy += v[1] / float64(len(p))
	}
	for i, v := range p {
		shadow[i] = [2]float64{v[0] + 2, v[1] + 8}
		inner[i] = [2]float64{cx + (v[0]-cx)*.81, cy + (v[1]-cy)*.81}
	}
	c.polygon(shadow, black, .8)
	c.outline(shadow, 17, black, .45)
	c.polygon(p, graphite, 1)
	c.outline(p, 17, black, 1)
	c.outline(p, 11, pink, .65)
	c.outline(p, 7, metal, 1)
	c.outline(p, 4, silver, .85)
	c.outline(p, 2, black, 1)
	c.polygon(inner, color.NRGBA{R: 12, G: 20, B: 26, A: 255}, 1)
	c.outline(inner, 1, cyan, .35)
	for i, v := range inner {
		x, y := cx+(v[0]-cx)*.65, cy+(v[1]-cy)*.65
		c.glowLine(x, y, cx, cy+float64(i-1)*11, 1.5, cyan, .7)
		c.ring(x, y, 3, 1.5, cyan, .8)
	}
	for _, v := range inner {
		c.screw(v[0], v[1], 5)
	}
}

func (c *canvas) screw(x, y, radius float64) {
	c.circle(x, y+1, radius+1.5, black, .85)
	c.circleGradient(x, y, radius, func(t float64) (color.NRGBA, float64) {
		return mix(silver, metal, t), 1
	})
	c.ring(x, y, radius, radius-.8, silver, .7)
	c.line(x-radius*.45, y+radius*.45, x+radius*.45, y-radius*.45, 1.5, black, 1)
}

func (c *canvas) relay(x, y, scale float64, col color.NRGBA) {
	for _, line := range [][4]float64{{-31, 0, -13, 0}, {-13, 0, 13, -14}, {15, 0, 31, 0}, {-31, 0, -31, -12}, {31, 0, 31, -12}} {
		c.glowLine(x+line[0]*scale, y+line[1]*scale, x+line[2]*scale, y+line[3]*scale, 2*scale, col, .9)
	}
	for _, dx := range []float64{-13, 15} {
		c.circle(x+dx*scale, y, 4.2*scale, graphite, 1)
		c.ring(x+dx*scale, y, 4.2*scale, 2.4*scale, col, 1)
	}
}

// Continuous outline lettering uses original authored paths rather than an
// external font, so the identity stays crisp and reproducible in every build.
func (c *canvas) wordmark(text string, x, y, unit float64, col color.NRGBA) {
	letters := map[rune][][][2]float64{
		'N': {{{0, 6}, {0, 0}, {4, 6}, {4, 0}}},
		'E': {{{4, 0}, {0, 0}, {0, 6}, {4, 6}}, {{0, 3}, {3.4, 3}}},
		'O': {{{.5, 0}, {3.5, 0}, {4, .5}, {4, 5.5}, {3.5, 6}, {.5, 6}, {0, 5.5}, {0, .5}, {.5, 0}}},
		'R': {{{0, 6}, {0, 0}, {3.5, 0}, {4, .5}, {4, 2.5}, {3.5, 3}, {0, 3}}, {{2, 3}, {4, 6}}},
		'L': {{{0, 0}, {0, 6}, {4, 6}}},
		'A': {{{0, 6}, {0, 1}, {1, 0}, {3, 0}, {4, 1}, {4, 6}}, {{0, 3}, {4, 3}}},
		'Y': {{{0, 0}, {2, 3}, {4, 0}}, {{2, 3}, {2, 6}}},
	}
	for _, ch := range text {
		for _, path := range letters[ch] {
			for i := 1; i < len(path); i++ {
				a, b := path[i-1], path[i]
				c.glowLine(x+a[0]*unit, y+a[1]*unit, x+b[0]*unit, y+b[1]*unit, unit*.4, col, .7)
				c.line(x+a[0]*unit, y+a[1]*unit, x+b[0]*unit, y+b[1]*unit, unit*.1, cyanWhite, .85)
			}
		}
		x += unit * 6
	}
}

func logoImage() image.Image {
	c := newCanvas(640, 200)
	c.wordmark("NEON", 155, 15, 15, cyan)
	c.wordmark("RELAY", 166, 122, 11, pink)
	return c.finish()
}

func ballImage() image.Image {
	c := newCanvas(int(table.BallFrame.Width), int(table.BallFrame.Height))
	c.circle(33, 36, 27, black, .4)
	c.circleGradient(32, 31, 27, func(t float64) (color.NRGBA, float64) {
		if t < .58 {
			return mix(silver, graphite, t/.58), 1
		}
		return mix(graphite, silver, (t-.58)/.42), 1
	})
	c.circle(25, 22, 13, silver, .85)
	c.circle(23, 19, 8, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, .95)
	c.line(19, 46, 31, 50, 2, pink, .8)
	c.line(47, 30, 48, 38, 2, cyan, .85)
	return c.finish()
}

func flipperImage() image.Image {
	c := newCanvas(int(table.FlipperFrame.Width), int(table.FlipperFrame.Height))
	shape := [][2]float64{{30, 13}, {149, 20}, {157, 25}, {160, 32}, {157, 40}, {149, 44}, {30, 51}}
	c.line(30, 36, 148, 36, 36, black, .65)
	c.polygon(shape, pink, 1)
	c.outline(shape, 3, color.NRGBA{R: 104, G: 18, B: 49, A: 255}, 1)
	c.circle(30, 32, 20, pink, 1)
	face := [][2]float64{{30, 17}, {148, 24}, {154, 28}, {155, 33}, {150, 39}, {30, 46}}
	c.polygon(face, color.NRGBA{R: 230, G: 230, B: 222, A: 255}, 1)
	c.outline(face, 1, silver, 1)
	c.line(46, 19, 145, 25, 2, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, 1)
	c.circle(30, 32, 17, graphite, 1)
	c.ring(30, 32, 16, 13, silver, 1)
	c.screw(30, 32, 10)
	return c.finish()
}

func bumperImage() image.Image {
	c := newCanvas(int(table.BumperFrame.Width), int(table.BumperFrame.Height))
	c.circle(65, 69, 58, black, .55)
	c.circleGradient(64, 64, 55, func(t float64) (color.NRGBA, float64) {
		return mix(graphite, metal, t), 1
	})
	c.ring(64, 64, 54, 52, silver, .8)
	c.circle(64, 64, 49, black, 1)
	c.ring(64, 64, 47, 40, pink, .85)
	c.ring(64, 64, 45, 43, cyanWhite, .75)
	for i := 0; i < 24; i++ {
		a := float64(i) * math.Pi / 12
		x, y := math.Cos(a), math.Sin(a)
		col := pink
		if i >= 12 && i < 18 {
			col = cyan
		}
		c.line(64+x*41, 64+y*41, 64+x*46, 64+y*46, 4, col, 1)
	}
	c.circleGradient(64, 62, 37, func(t float64) (color.NRGBA, float64) {
		return mix(color.NRGBA{R: 36, G: 57, B: 66, A: 255}, black, t), 1
	})
	c.ring(64, 62, 37, 35.5, metal, 1)
	c.ring(64, 62, 32, 31, cyan, .6)
	c.relay(64, 66, .72, cyan)
	c.line(46, 38, 65, 34, 2, silver, .45)
	for _, a := range []float64{math.Pi / 2, math.Pi * 7 / 6, math.Pi * 11 / 6} {
		c.screw(64+math.Cos(a)*52, 64+math.Sin(a)*52, 4)
	}
	return c.finish()
}

func postImage() image.Image {
	c := newCanvas(int(table.PostFrame.Width), int(table.PostFrame.Height))
	c.circle(25, 27, 20, black, .6)
	c.circle(24, 24, 19, graphite, 1)
	c.ring(24, 24, 18, 14, metal, 1)
	c.ring(24, 24, 14, 10, cyan, .85)
	c.screw(24, 23, 9)
	return c.finish()
}

func targetImage() image.Image     { return targetFace(false) }
func targetDownImage() image.Image { return targetFace(true) }

func targetFace(down bool) image.Image {
	c := newCanvas(int(table.TargetFrame.Width), int(table.TargetFrame.Height))
	c.roundedRect(19, 8, 49, 92, 10, black, .7)
	c.roundedRect(18, 4, 46, 88, 8, metal, 1)
	c.roundedRect(20, 6, 44, 86, 6, black, 1)
	col, opacity := pink, 1.0
	if down {
		col, opacity = metal, .35
	}
	c.roundedRect(21, 12, 43, 80, 5, col, .5*opacity)
	c.roundedRect(23, 14, 41, 78, 4, graphite, 1)
	for i, length := range []float64{6, 13, 17, 10} {
		y := 29.0 + float64(i)*11
		c.glowLine(32-length/2, y, 32+length/2, y, 3, col, opacity)
	}
	c.line(23, 8, 41, 8, 1.5, silver, .7)
	return c.finish()
}

func laneLightImage() image.Image    { return laneLight(true) }
func laneLightOffImage() image.Image { return laneLight(false) }

func laneLight(lit bool) image.Image {
	c := newCanvas(48, 96)
	c.roundedRect(8, 4, 40, 92, 15, metal, 1)
	c.roundedRect(10, 6, 38, 90, 13, black, 1)
	col, opacity := metal, .4
	if lit {
		col, opacity = cyan, 1
	}
	for _, y := range []float64{28, 47, 66} {
		c.glowLine(17, y+5, 24, y-3, 3, col, opacity)
		c.glowLine(24, y-3, 31, y+5, 3, col, opacity)
	}
	c.circle(24, 81, 3, gold, opacity)
	return c.finish()
}

func plungerImage() image.Image {
	c := newCanvas(56, 180)
	c.roundedRect(19, 5, 37, 154, 8, black, 1)
	c.line(28, 10, 28, 155, 8, metal, 1)
	c.line(26, 12, 26, 153, 2, silver, .9)
	for y := 48.0; y < 136; y += 9 {
		c.line(17, y+3, 39, y-2, 5, black, 1)
		c.line(17, y+2, 39, y-3, 3, silver, .9)
	}
	c.roundedRect(10, 144, 46, 175, 9, black, 1)
	c.roundedRect(12, 145, 44, 172, 8, pink, .95)
	c.line(18, 148, 38, 148, 2, silver, .8)
	return c.finish()
}
