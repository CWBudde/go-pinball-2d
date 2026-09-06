//go:build js && wasm

package platform

import (
	"syscall/js"
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/table"
	"github.com/gonutz/prototype/draw"
)

func recordCanvasStrokes(t *testing.T) js.Value {
	t.Helper()
	previousCanvas, previousStrokeCanvas, previousContext := displayCanvas, strokeCanvas, strokeContext
	ctx := js.Global().Get("Function").New(`
		const ctx = {calls: [], strokes: [], stack: [], lineWidth: 1,
			strokeStyle: 'original', lineCap: 'round', lineJoin: 'bevel'};
		for (const name of ['save', 'restore', 'beginPath', 'moveTo', 'lineTo', 'ellipse', 'stroke', 'strokeRect']) {
			ctx[name] = function(...args) {
				this.calls.push([name, ...args]);
				if (name === 'save') this.stack.push([this.lineWidth, this.strokeStyle, this.lineCap, this.lineJoin]);
				if (name === 'restore') [this.lineWidth, this.strokeStyle, this.lineCap, this.lineJoin] = this.stack.pop();
				if (name === 'stroke' || name === 'strokeRect') this.strokes.push([this.lineWidth, this.strokeStyle, this.lineCap, this.lineJoin]);
			};
		}
		return ctx;
	`).Invoke()
	displayCanvas = js.Global().Get("Object").New()
	getContext := js.FuncOf(func(js.Value, []js.Value) any { return ctx })
	displayCanvas.Set("getContext", getContext)
	strokeCanvas, strokeContext = js.Undefined(), js.Undefined()
	t.Cleanup(func() {
		displayCanvas, strokeCanvas, strokeContext = previousCanvas, previousStrokeCanvas, previousContext
		getContext.Release()
	})
	return ctx
}

func TestCanvasStrokeGeometryAndState(t *testing.T) {
	for _, test := range []struct {
		name string
		draw func()
		want []string
	}{
		{
			"line", func() {
				new(renderer).thickLine(nil, newViewport(1440, 2160), table.New().OuterWalls[0].Segment, draw.Red, 3)
			},
			[]string{`["save"]`, `["beginPath"]`, `["moveTo",80.5,300.5]`, `["lineTo",110.5,210.5]`, `["stroke"]`, `["restore"]`},
		},
		{
			"ellipse", func() { new(renderer).thickEllipse(nil, 10, 20, 26, 12, draw.Red, 7) },
			[]string{`["save"]`, `["beginPath"]`, `["ellipse",23,26,12.5,5.5,0,0,6.283185307179586]`, `["stroke"]`, `["restore"]`},
		},
		{
			"rectangle", func() { new(renderer).thickRect(nil, 10, 20, 260, 16, draw.Red, 7) },
			[]string{`["save"]`, `["strokeRect",10.5,20.5,259,15]`, `["restore"]`},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := recordCanvasStrokes(t)
			test.draw()
			calls := ctx.Get("calls")
			if calls.Length() != len(test.want) {
				t.Fatalf("Canvas calls = %d, want %d", calls.Length(), len(test.want))
			}
			for i, want := range test.want {
				if got := js.Global().Get("JSON").Call("stringify", calls.Index(i)).String(); got != want {
					t.Errorf("call %d = %s, want %s", i, got, want)
				}
			}
			stroke := ctx.Get("strokes").Index(0)
			if stroke.Index(0).Int() != 7 || stroke.Index(1).String() != "rgba(255,0,0,1.000000)" || stroke.Index(2).String() != "butt" || stroke.Index(3).String() != "miter" {
				t.Fatalf("unexpected stroke state: %s", js.Global().Get("JSON").Call("stringify", stroke))
			}
			if ctx.Get("lineWidth").Int() != 1 || ctx.Get("strokeStyle").String() != "original" || ctx.Get("lineCap").String() != "round" || ctx.Get("lineJoin").String() != "bevel" || ctx.Get("stack").Length() != 0 {
				t.Fatal("stroke changed the shared Canvas context state")
			}
		})
	}
}

// Unexpected use of draw.Window's per-pixel primitives panics through the nil
// embedded interface, so this also protects against falling back on the web.
type canvasTableWindow struct{ draw.Window }

func (canvasTableWindow) DrawImageFileTo(string, int, int, int, int, int) error { return nil }

func TestCanvasTableStrokesHaveConstantCallCount(t *testing.T) {
	current := game.New(table.New(), nil)
	for _, scale := range []int{1, 2, 3} {
		ctx := recordCanvasStrokes(t)
		view := newViewport(720*scale, 1080*scale)
		new(renderer).drawTable(canvasTableWindow{}, current, view)
		// 27 wall/slingshot lines and 3 rollover ellipses, regardless of DPR.
		if got := ctx.Get("strokes").Length(); got != 30 {
			t.Fatalf("scale %d: %d strokes, want 30", scale, got)
		}
		if got := ctx.Get("calls").Length(); got != 27*6+3*5 {
			t.Fatalf("scale %d: %d Canvas calls, want 177", scale, got)
		}
	}
}
