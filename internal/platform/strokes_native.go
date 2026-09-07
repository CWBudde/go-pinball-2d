//go:build !js || !wasm

package platform

import "github.com/gonutz/prototype/draw"

func drawCanvasText(_ string, _, _ int, _ float32, _ draw.Color, _ bool) bool {
	return false
}

func drawCanvasLine(_, _, _, _ int, _ draw.Color, _ int) bool { return false }

func drawCanvasEllipse(_, _, _, _ int, _ draw.Color, _ int) bool { return false }

func drawCanvasRect(_, _, _, _ int, _ draw.Color, _ int) bool { return false }
