//go:build js && wasm

package platform

import (
	"math"
	"syscall/js"
)

var displayCanvas js.Value

// prepareCanvas runs after draw initializes the canvas and before each frame.
// draw also resets its dimensions and inline styles on fullscreen transitions.
func prepareCanvas() {
	if !displayCanvas.Truthy() {
		displayCanvas = js.Global().Get("document").Call("getElementById", "gameCanvas")
	}
	if !displayCanvas.Truthy() {
		return
	}
	resizeCanvas(displayCanvas)
}

func resizeCanvas(canvas js.Value) {
	style := canvas.Get("style")
	width, height := "100%", "100%"
	if js.Global().Get("document").Get("fullscreenElement").Equal(canvas) {
		width, height = "100vw", "100vh"
	}
	if style.Get("width").String() != width {
		style.Set("width", width)
	}
	if style.Get("height").String() != height {
		style.Set("height", height)
	}

	ratio := 1.0
	if value := js.Global().Get("devicePixelRatio"); value.Type() == js.TypeNumber {
		if candidate := value.Float(); candidate > 0 && !math.IsNaN(candidate) && !math.IsInf(candidate, 0) {
			ratio = candidate
		}
	}
	bounds := canvas.Call("getBoundingClientRect")
	for _, dimension := range []string{"width", "height"} {
		pixels := max(1, int(math.Round(bounds.Get(dimension).Float()*ratio)))
		// Assigning even an unchanged dimension clears the canvas and context.
		if canvas.Get(dimension).Int() != pixels {
			canvas.Set(dimension, pixels)
		}
	}
}
