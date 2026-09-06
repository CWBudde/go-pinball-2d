//go:build !js || !wasm

package platform

import "github.com/gonutz/prototype/draw"

func drawCanvasLine(_, _, _, _ int, _ draw.Color, _ int) bool { return false }

func drawCanvasEllipse(_, _, _, _ int, _ draw.Color, _ int) bool { return false }

func drawCanvasRect(_, _, _, _ int, _ draw.Color, _ int) bool { return false }
