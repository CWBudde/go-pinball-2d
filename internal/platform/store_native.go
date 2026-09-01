//go:build !js || !wasm

package platform

import "github.com/CWBudde/go-pinball-2d/internal/game"

func newHighScoreStore() game.HighScoreStore { return &game.MemoryStore{} }
