package main

import (
	"image"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
)

// Static layers use logical table coordinates; the same line paths produce
// visible contact edges and physics. Shadows and printed panels have no height.
func tableShadowImage() image.Image {
	c := newCanvas(int(table.Width), int(table.Height))
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
	return c.finish()
}

func tableHardwareImage() image.Image {
	c := newCanvas(int(table.Width), int(table.Height))
	d := table.New()
	// Low-profile upper routing plate: a printed backing below the ball, with
	// independently colliding guide rails on top. The central launch orbit is open.
	panel := [][2]float64{{111, 251}, {99, 199}, {120, 153}, {168, 120}, {250, 120}, {283, 106}, {432, 106}, {464, 120}, {546, 120}, {588, 153}, {602, 199}, {584, 251}, {564, 251}, {571, 195}, {539, 153}, {451, 153}, {422, 132}, {293, 132}, {263, 153}, {180, 153}, {142, 183}, {139, 251}}
	c.polygon(panel, graphite, 1)
	c.outline(panel, 2, metal, .65)
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
	// Drain aprons occupy the sealed regions outside the lower playfield walls.
	for _, panel := range [][][2]float64{
		{{52, 861}, {144, 1021}, {249, 1058}, {53, 1058}},
		{{615, 850}, {520, 1019}, {520, 1058}, {615, 1058}},
	} {
		c.polygon(panel, graphite, 1)
		c.outline(panel, 2, metal, .6)
	}
	return c.finish()
}

// Foreground is restricted to the cabinet and the lip below the drain sensor.
// It cannot hide a live ball, flipper, or target in the playable area.
func tableForegroundImage() image.Image {
	c := newCanvas(int(table.Width), int(table.Height))
	// Draw perimeter bands as paths so their interior remains transparent.
	frame := [][2]float64{{31, 17}, {688, 17}, {703, 32}, {703, 1049}, {688, 1064}, {31, 1064}, {17, 1049}, {17, 32}}
	c.outline(frame, 24, black, 1)
	c.outline(frame, 20, graphite, 1)
	c.outline(frame, 2, metal, 1)
	for _, p := range [][2]float64{{20, 21}, {700, 21}, {20, 1059}, {700, 1059}} {
		c.screw(p[0], p[1], 5)
	}
	for _, bounds := range [][4]float64{{42, 12, 373, 49}, {438, 12, 675, 49}} {
		c.roundedRect(bounds[0], bounds[1], bounds[2], bounds[3], 7, metal, .85)
		c.roundedRect(bounds[0]+1, bounds[1]+1, bounds[2]-1, bounds[3]-1, 6, black, 1)
		c.line(bounds[0]+9, bounds[3]-2, bounds[2]-9, bounds[3]-2, 1, cyan, .25)
	}
	c.line(264, 1073, 506, 1073, 10, graphite, 1)
	c.line(264, 1069, 506, 1069, 2, metal, .8)
	return c.finish()
}
