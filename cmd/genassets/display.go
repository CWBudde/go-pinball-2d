package main

import (
	"image"
	"image/color"

	"github.com/CWBudde/go-pinball-2d/internal/display"
)

func fontAtlasImage() image.Image {
	c := newCanvas(display.Columns*display.CellWidth, display.Rows*display.CellHeight*display.Colors)
	palette := []color.NRGBA{mix(cyan, cyanWhite, .6), silver, gold, pink, materialRGB(75, 116, 125)}
	for tone, col := range palette {
		for _, ch := range display.Characters {
			x, y := display.Cell(ch, tone)
			for _, path := range display.Paths(ch) {
				for i := 1; i < len(path); i++ {
					a, b := path[i-1], path[i]
					c.line(float64(x)+4+a[0]*6, float64(y)+6+a[1]*6, float64(x)+4+b[0]*6, float64(y)+6+b[1]*6, 3.8, col, 1)
				}
			}
		}
	}
	img := c.finishAlpha()
	// GLFW generates mipmaps from straight-alpha RGBA. Preserve the ink RGB
	// throughout each color strip, including transparent padding, so filtered
	// edges lose coverage without also darkening their color. Premultiplied
	// software/Canvas compositing still sees the same glyph coverage.
	for y := range img.Bounds().Dy() {
		col := palette[y/(display.Rows*display.CellHeight)]
		for x := range img.Bounds().Dx() {
			i := img.PixOffset(x, y)
			img.Pix[i], img.Pix[i+1], img.Pix[i+2] = col.R, col.G, col.B
		}
	}
	return img
}

// Flush glass instruments are baked below the ball. They have no contact edge.
func (c *canvas) instrument(x0, y0, x1, y1 float64) {
	c.roundedRect(x0-2, y0-1, x1+2, y1+3, 7, black, .8)
	c.roundedRect(x0, y0, x1, y1, 6, metal, 1)
	c.roundedRect(x0+.8, y0+.8, x1-.8, y1-.8, 5, silver, .6)
	c.roundedRect(x0+2, y0+2, x1-2, y1-2, 4, black, 1)
	c.roundedRect(x0+3, y0+3, x1-3, y0+8, 3, materialRGB(33, 54, 63), .6)
	c.line(x0+8, y1-3, x1-8, y1-3, .7, cyan, .3)
	// Dim rows suggest smoked display glass, without competing with glyphs.
	for y := y0 + 10; y < y1-4; y += 3 {
		c.line(x0+5, y, x1-5, y, .35, metal, .08)
	}
}

func (c *canvas) printedText(text string, x, y, height float64, col color.NRGBA) {
	unit := height / 6
	for _, ch := range text {
		for _, path := range display.Paths(ch) {
			for i := 1; i < len(path); i++ {
				a, b := path[i-1], path[i]
				c.line(x+a[0]*unit, y+a[1]*unit, x+b[0]*unit, y+b[1]*unit, unit*.65, col, .85)
			}
		}
		x += unit * 6
	}
}
