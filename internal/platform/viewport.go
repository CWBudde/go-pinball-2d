package platform

import (
	"math"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
)

type viewport struct {
	scale   float64
	offsetX int
	offsetY int
	width   int
	height  int
}

func newViewport(screenWidth, screenHeight int) viewport {
	if screenWidth < 1 {
		screenWidth = 1
	}
	if screenHeight < 1 {
		screenHeight = 1
	}
	scale := math.Min(float64(screenWidth)/table.Width, float64(screenHeight)/table.Height)
	width := int(math.Round(table.Width * scale))
	height := int(math.Round(table.Height * scale))
	return viewport{
		scale: scale, width: width, height: height,
		offsetX: (screenWidth - width) / 2, offsetY: (screenHeight - height) / 2,
	}
}

func (v viewport) point(point physics.Vec) (int, int) {
	return v.x(point.X), v.y(point.Y)
}

func (v viewport) x(value float64) int { return v.offsetX + int(math.Round(value*v.scale)) }
func (v viewport) y(value float64) int { return v.offsetY + int(math.Round(value*v.scale)) }
func (v viewport) size(value float64) int {
	return max(1, int(math.Round(value*v.scale)))
}
