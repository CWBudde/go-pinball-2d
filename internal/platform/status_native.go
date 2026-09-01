//go:build !js || !wasm

package platform

import "github.com/CWBudde/go-pinball-2d/internal/game"

func publishStatus(*game.Game) {}
