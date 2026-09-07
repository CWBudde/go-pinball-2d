package main

import (
	"image"
	"image/color"
	"math"

	"github.com/CWBudde/go-pinball-2d/internal/display"
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
	c := materialCanvas(720, 1080)
	c.verticalGradient(graphite, materialRGB(10, 15, 19))

	// Subtle etched grain, with no free-floating stars or space artwork.
	seed := rng(0x52454c4159)
	for y := 42.0; y < 1040; y += 3 {
		for x := 43.0; x < 680; x += 4 {
			c.line(x, y, x+2, y, .5, silver, .015+seed.unit()*.022)
		}
	}
	c.boardDetails()
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

	for i := 0; i < 9; i++ {
		y := 305.0 + float64(i)*69
		c.circle(643, y, 4, black, 1)
		c.ring(643, y, 3, 1.7, gold, .7)
		if i < 8 {
			c.line(643, y+6, 643, y+62, 1, metal, .45)
		}
	}
	return c.finishAlpha()
}

func playfieldMarkingsImage() image.Image {
	c := materialCanvas(720, 1080)
	// Screen-printed center identity and a real switching-contact diagram.
	c.wordmark("NEON", table.TitleOrigin.X, table.TitleOrigin.Y, 8.0, cyan)
	c.wordmark("RELAY", table.TitleOrigin.X-4, table.TitleOrigin.Y+55, 6.4, pink)
	c.line(294, 557, 314, 557, 5, black, 1)
	c.line(314, 557, 364, 543, 5, metal, 1)
	c.glowLine(294, 555, 314, 555, 2, cyan, .8)
	c.glowLine(314, 555, 364, 541, 2, cyan, .8)
	c.glowLine(360, 555, 396, 555, 2, pink, .8)
	for _, x := range []float64{314, 360} {
		c.circle(x, 555, 5, black, 1)
		c.ring(x, 555, 4.5, 2.8, silver, .9)
		c.ring(x, 555, 3.8, 2.8, cyan, .8)
	}
	// Printed legends stay in dedicated flush areas, below live balls.
	for i, x := range []float64{239, 359, 479} {
		c.instrument(x-23, 229, x+23, 252)
		c.printedText([]string{"01", "02", "03"}[i], x-11, 235, 10, silver)
	}
	c.instrument(453, 651, 569, 699)
	c.printedText("RELAY BANK", 465, 660, 9, pink)
	for i := 0; i < 3; i++ {
		y := 764.0 + float64(i)*23
		c.glowLine(table.PlayfieldCenter-19, y+11, table.PlayfieldCenter, y, 2, pink, .55)
		c.glowLine(table.PlayfieldCenter, y, table.PlayfieldCenter+19, y+11, 2, pink, .55)
	}

	// Flush directional inserts identify the two real channels on either side.
	// Their positions come from routing metadata, never a second visual layout.
	d := table.New()
	for i, lanes := range [][]table.Lane{d.Inlanes, d.Outlanes} {
		col := cyan
		if i == 1 {
			col = gold
		}
		for _, lane := range lanes {
			p := lane.Segment.A.Add(lane.Segment.B).Mul(.5)
			c.roundedRect(p.X-8, p.Y-18, p.X+8, p.Y+18, 5, metal, .7)
			c.roundedRect(p.X-7, p.Y-17, p.X+7, p.Y+17, 4, black, 1)
			for _, y := range []float64{p.Y - 7, p.Y + 3} {
				c.line(p.X-4, y, p.X, y+4, 1.5, col, .8)
				c.line(p.X, y+4, p.X+4, y, 1.5, col, .8)
			}
		}
	}

	return c.finishAlpha()
}

func (c *canvas) outline(p [][2]float64, width float64, col color.NRGBA, opacity float64) {
	for i, a := range p {
		b := p[(i+1)%len(p)]
		c.line(a[0], a[1], b[0], b[1], width, col, opacity)
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
	// Complete each material pass before the next, preserving continuous joins.
	for _, layer := range []struct {
		dx, dy, width float64
		col           color.NRGBA
		glow          bool
	}{
		{1, 2, .7, black, false},
		{0, 0, .58, metal, false},
		{0, 0, .3, col, true},
		{-.3, -.5, .09, cyanWhite, false},
	} {
		cursor := x
		for _, ch := range text {
			for _, path := range display.Paths(ch) {
				for i := 1; i < len(path); i++ {
					a, b := path[i-1], path[i]
					ax, ay := cursor+a[0]*unit+layer.dx, y+a[1]*unit+layer.dy
					bx, by := cursor+b[0]*unit+layer.dx, y+b[1]*unit+layer.dy
					if layer.glow {
						c.glowLine(ax, ay, bx, by, unit*layer.width, layer.col, .85)
					} else {
						c.line(ax, ay, bx, by, unit*layer.width, layer.col, .9)
					}
				}
			}
			cursor += unit * 6
		}
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
	c.ellipseSurface(32, 31, 27, 27, func(x, y, r float64) (color.NRGBA, float64) {
		// A spherical chrome reflection: broad cool key, dark horizon, sharp strip
		// and grazing LED color. No painted shadow inside the physical ball image.
		z := math.Sqrt(math.Max(0, 1-r*r))
		key := math.Max(0, -.55*x-.65*y+.52*z)
		v := 24 + 115*bell(key, .87, .16) + 100*bell(y+.2*x, -.43, .11)
		v += 95*math.Pow(r, 9) + 50*bell(y, .63, .06)
		v -= 18 * bell(y, .22, .14)
		cyanEdge := 45 * bell(x, -.86, .12)
		pinkEdge := 55 * bell(x, .8, .12) * math.Max(0, y+.3)
		return materialRGB(v+pinkEdge, v*1.06+cyanEdge, v*1.13+cyanEdge+pinkEdge*.3), 1
	})
	return c.finishAlpha()
}

func ballShadowImage() image.Image {
	c := materialCanvas(48, 32)
	c.ellipseSurface(24, 16, 22, 13, func(_, _, r float64) (color.NRGBA, float64) {
		return black, .7 * math.Pow(math.Max(0, 1-r*r), 1.1)
	})
	return c.finishAlpha()
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
