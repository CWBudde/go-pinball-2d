//go:build js && wasm

package platform

import (
	"fmt"
	"math"
	"syscall/js"

	"github.com/gonutz/prototype/draw"
)

var strokeCanvas, strokeContext js.Value

// Canvas paths avoid prototype/draw's per-pixel Go-to-JavaScript calls for
// lines and ellipses. Keep the context's state intact for its other primitives.
func beginCanvasStroke(color draw.Color, width int) js.Value {
	if !displayCanvas.Truthy() {
		return js.Undefined()
	}
	if !strokeCanvas.Equal(displayCanvas) {
		strokeCanvas = displayCanvas
		strokeContext = displayCanvas.Call("getContext", "2d")
	}
	if !strokeContext.Truthy() {
		return js.Undefined()
	}
	ctx := strokeContext
	ctx.Call("save")
	ctx.Set("strokeStyle", fmt.Sprintf("rgba(%d,%d,%d,%f)", int(color.R*255), int(color.G*255), int(color.B*255), color.A))
	ctx.Set("lineWidth", width)
	ctx.Set("lineCap", "butt")
	ctx.Set("lineJoin", "miter")
	return ctx
}

func drawCanvasLine(ax, ay, bx, by int, color draw.Color, width int) bool {
	ctx := beginCanvasStroke(color, width)
	if !ctx.Truthy() {
		return false
	}
	ctx.Call("beginPath")
	ctx.Call("moveTo", float64(ax)+.5, float64(ay)+.5)
	ctx.Call("lineTo", float64(bx)+.5, float64(by)+.5)
	ctx.Call("stroke")
	ctx.Call("restore")
	return true
}

func drawCanvasEllipse(x, y, width, height int, color draw.Color, strokeWidth int) bool {
	if width <= 0 || height <= 0 {
		return true
	}
	ctx := beginCanvasStroke(color, strokeWidth)
	if !ctx.Truthy() {
		return false
	}
	ctx.Call("beginPath")
	ctx.Call("ellipse", float64(x)+float64(width)/2, float64(y)+float64(height)/2,
		float64(width-1)/2, float64(height-1)/2, 0, 0, 2*math.Pi)
	ctx.Call("stroke")
	ctx.Call("restore")
	return true
}

func drawCanvasRect(x, y, width, height int, color draw.Color, strokeWidth int) bool {
	if width <= 0 || height <= 0 {
		return true
	}
	ctx := beginCanvasStroke(color, strokeWidth)
	if !ctx.Truthy() {
		return false
	}
	ctx.Call("strokeRect", float64(x)+.5, float64(y)+.5, width-1, height-1)
	ctx.Call("restore")
	return true
}

// Use a system monospace face for the instrument labels and score displays.
func drawCanvasText(text string, x, y int, scale float32, color draw.Color, centered bool) bool {
	ctx := beginCanvasStroke(color, 1)
	if !ctx.Truthy() {
		return false
	}
	ctx.Set("font", fmt.Sprintf("500 %.2fpx ui-monospace, monospace", 24*scale))
	ctx.Set("fillStyle", ctx.Get("strokeStyle"))
	ctx.Set("textBaseline", "top")
	ctx.Set("textAlign", "left")
	if centered {
		ctx.Set("textAlign", "center")
	}
	ctx.Call("fillText", text, x, y)
	ctx.Call("restore")
	return true
}
