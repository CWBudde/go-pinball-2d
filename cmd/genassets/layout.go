package main

import (
	"image"
	"image/color"
	"math"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
)

// Static layers use logical table coordinates; the same line paths produce
// visible contact edges and physics. Shadows and printed panels have no height.
func tableShadowImage() image.Image {
	c := materialCanvas(int(table.Width), int(table.Height))
	d := table.New()
	for _, group := range [][]physics.LineCollider{d.OuterWalls, d.ShooterLane, d.GuideWalls} {
		for _, wall := range group {
			a, b := wall.Segment.A, wall.Segment.B
			c.line(a.X+3, a.Y+7, b.X+3, b.Y+7, wall.Radius*2+12, black, .4)
		}
	}
	for _, sling := range d.Slingshots {
		points := make([][2]float64, 3)
		for i, p := range sling.Triangle {
			points[i] = [2]float64{p.X + 3, p.Y + 8}
		}
		c.polygon(points, black, .65)
		c.outline(points, 19, black, .3)
	}
	return c.finishAlpha()
}

func tableHardwareImage() image.Image {
	c := materialCanvas(int(table.Width), int(table.Height))
	d := table.New()
	// Low-profile upper routing plate: a printed backing below the ball, with
	// independently colliding guide rails on top. The central launch orbit is open.
	panel := [][2]float64{{111, 251}, {99, 199}, {120, 153}, {168, 120}, {250, 120}, {283, 106}, {432, 106}, {464, 120}, {546, 120}, {588, 153}, {602, 199}, {584, 251}, {564, 251}, {571, 195}, {539, 153}, {451, 153}, {422, 132}, {293, 132}, {263, 153}, {180, 153}, {142, 183}, {139, 251}}
	c.polygon(panel, graphite, 1)
	c.outline(panel, 2, metal, .65)
	c.outline(panel, .6, silver, .55)
	// These etched channels are flush board details under the launch orbit.
	for _, x := range []float64{165, 552} {
		for i := 0; i < 5; i++ {
			y := 168 + float64(i)*12
			c.line(x-4, y, x+4, y, 1, metal, .65)
		}
	}
	c.relay(357, 120, .35, gold)
	for _, walls := range [][]physics.LineCollider{d.OuterWalls, d.ShooterLane, d.GuideWalls} {
		c.rails(walls)
	}
	for _, sling := range d.Slingshots {
		points := make([][2]float64, 3)
		for i, p := range sling.Triangle {
			points[i] = [2]float64{p.X, p.Y}
		}
		c.plastic(points)
	}
	for _, x := range []float64{188, 300, 420, 532} {
		c.screw(x, 148, 5)
		c.screw(x, 213, 5)
	}
	// A recessed plunger well has no raised cover over the shooter's ball path.
	c.roundedRect(636, 905, 675, 1038, 14, black, 1)
	c.roundedRect(638, 907, 673, 1036, 12, metal, .7)
	c.roundedRect(640, 909, 671, 1034, 10, black, 1)
	c.line(642, 923, 642, 1024, 1, silver, .6)
	c.line(669, 923, 669, 1024, 1, pink, .3)
	for y := 918.0; y < 977; y += 8 {
		c.line(662, y, 666, y, .7, metal, .65)
	}
	// Drain aprons occupy the sealed regions outside the lower playfield walls.
	for _, side := range []int{0, 1} {
		panel := [][2]float64{{52, 861}, {144, 1021}, {231, 1058}, {53, 1058}}
		if side == 1 {
			for i := range panel {
				panel[i][0] = 2*table.PlayfieldCenter - panel[i][0]
			}
		}
		c.polygon(panel, graphite, 1)
		c.outline(panel, 2, metal, .6)
		// Inset ribs and machining marks stay on the sealed apron side of the wall.
		cx, cy := 0.0, 0.0
		for _, p := range panel {
			cx += p[0] / float64(len(panel))
			cy += p[1] / float64(len(panel))
		}
		inner := make([][2]float64, len(panel))
		for i, p := range panel {
			inner[i] = [2]float64{cx + (p[0]-cx)*.8, cy + (p[1]-cy)*.8}
		}
		c.outline(inner, 1, black, 1)
		for i := 0; i < 4; i++ {
			y := cy + float64(i)*5
			c.line(cx-11, y, cx+11, y, 1, metal, .45)
		}
		c.bumperFastener(cx, cy+35)
	}
	return c.finishAlpha()
}

// Foreground is restricted to the cabinet and the lip below the drain sensor.
// It cannot hide a live ball, flipper, or target in the playable area.
func tableForegroundImage() image.Image {
	c := materialCanvas(int(table.Width), int(table.Height))
	// Rounded cabinet lip stays outside the existing perimeter colliders.
	frame := roundedPanel([][2]float64{{32, 18}, {688, 18}, {702, 32}, {702, 1048}, {688, 1062}, {32, 1062}, {18, 1048}, {18, 32}}, 10)
	for _, layer := range []struct {
		width float64
		col   color.NRGBA
	}{
		{28, black}, {24, metal}, {22, silver}, {20, black}, {17, graphite}, {1.1, metal},
	} {
		c.outline(frame, layer.width, layer.col, 1)
	}
	// Restrained brushed wear on the broad cabinet surface.
	for y := 70.0; y < 1030; y += 7 {
		c.line(12, y, 22, y-2, .45, silver, .12)
		c.line(698, y, 708, y-2, .45, silver, .10)
	}

	for _, p := range [][2]float64{{20, 21}, {700, 21}, {20, 1059}, {700, 1059}} {
		c.bumperFastener(p[0], p[1])
	}
	for _, bounds := range [][4]float64{{42, 12, 373, 49}, {438, 12, 675, 49}} {
		c.roundedRect(bounds[0]-3, bounds[1]-3, bounds[2]+3, bounds[3]+3, 9, black, 1)
		c.roundedRect(bounds[0]-1, bounds[1]-1, bounds[2]+1, bounds[3]+1, 8, silver, .9)
		c.roundedRect(bounds[0], bounds[1], bounds[2], bounds[3], 7, metal, .85)
		c.roundedRect(bounds[0]+1, bounds[1]+1, bounds[2]-1, bounds[3]-1, 6, black, 1)
		c.line(bounds[0]+9, bounds[3]-2, bounds[2]-9, bounds[3]-2, 1, cyan, .25)
	}
	c.line(table.PlayfieldCenter-121, 1073, table.PlayfieldCenter+121, 1073, 10, graphite, 1)
	c.line(table.PlayfieldCenter-121, 1069, table.PlayfieldCenter+121, 1069, 2, metal, .8)
	return c.finishAlpha()
}

// Rounded decorative panels never define ball travel. Contact curves come only
// from table.New, including the upper arch's sampled quadratic paths.
func roundedPanel(p [][2]float64, radius float64) [][2]float64 {
	var out [][2]float64
	for i, v := range p {
		prev, next := p[(i+len(p)-1)%len(p)], p[(i+1)%len(p)]
		toward := func(q [2]float64) [2]float64 {
			dx, dy := q[0]-v[0], q[1]-v[1]
			scale := math.Min(.4, radius/math.Hypot(dx, dy))
			return [2]float64{v[0] + dx*scale, v[1] + dy*scale}
		}
		a, b := toward(prev), toward(next)
		for j := 0; j <= 6; j++ {
			t := float64(j) / 6
			u := 1 - t
			out = append(out, [2]float64{u*u*a[0] + 2*u*t*v[0] + t*t*b[0], u*u*a[1] + 2*u*t*v[1] + t*t*b[1]})
		}
	}
	return out
}
