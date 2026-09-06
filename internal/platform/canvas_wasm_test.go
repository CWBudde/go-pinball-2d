//go:build js && wasm

package platform

import (
	"syscall/js"
	"testing"
)

func TestCanvasResolutionFollowsDisplay(t *testing.T) {
	global := js.Global()
	previousDocument, previousRatio := global.Get("document"), global.Get("devicePixelRatio")
	t.Cleanup(func() {
		global.Set("document", previousDocument)
		global.Set("devicePixelRatio", previousRatio)
	})
	document := global.Get("Object").New()
	global.Set("document", document)
	canvas := global.Get("Object").New()
	style := global.Get("Object").New()
	canvas.Set("style", style)
	var cssWidth, cssHeight float64
	bounds := js.FuncOf(func(js.Value, []js.Value) any {
		return map[string]any{"width": cssWidth, "height": cssHeight}
	})
	defer bounds.Release()
	canvas.Set("getBoundingClientRect", bounds)

	for _, test := range []struct {
		name                  string
		width, height         float64
		ratio                 any
		fullscreen            bool
		wantWidth, wantHeight int
	}{
		{"startup after renderer reset", 720, 1080, 2, false, 1440, 2160},
		{"fullscreen after renderer reset", 1920, 1080, 2, true, 3840, 2160},
		{"exit fullscreen into cabinet", 360, 540, 2, false, 720, 1080},
		{"resize and fractional DPR", 500, 750, 1.25, false, 625, 938},
		{"DPR change", 500, 750, 1, false, 500, 750},
		{"missing DPR", 500, 750, js.Undefined(), false, 500, 750},
	} {
		t.Run(test.name, func(t *testing.T) {
			cssWidth, cssHeight = test.width, test.height
			global.Set("devicePixelRatio", test.ratio)
			document.Set("fullscreenElement", js.Null())
			if test.fullscreen {
				document.Set("fullscreenElement", canvas)
			}
			// Simulate draw's startup/fullscreen dimension and style resets.
			canvas.Set("width", 720)
			canvas.Set("height", 1080)
			style.Set("width", "720px")
			style.Set("height", "1080px")
			resizeCanvas(canvas)
			if gotWidth, gotHeight := canvas.Get("width").Int(), canvas.Get("height").Int(); gotWidth != test.wantWidth || gotHeight != test.wantHeight {
				t.Fatalf("backing resolution = %dx%d, want %dx%d", gotWidth, gotHeight, test.wantWidth, test.wantHeight)
			}
			wantWidth, wantHeight := "100%", "100%"
			if test.fullscreen {
				wantWidth, wantHeight = "100vw", "100vh"
			}
			if style.Get("width").String() != wantWidth || style.Get("height").String() != wantHeight {
				t.Fatalf("canvas CSS = %s by %s, want %s by %s", style.Get("width"), style.Get("height"), wantWidth, wantHeight)
			}
		})
	}
}
