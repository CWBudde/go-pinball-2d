//go:build js && wasm

package platform

import (
	"fmt"
	"strconv"
	"syscall/js"

	"github.com/CWBudde/go-pinball-2d/internal/game"
)

func publishStatus(current *game.Game) {
	canvas := js.Global().Get("document").Call("getElementById", "gameCanvas")
	if !canvas.Truthy() {
		return
	}
	dataset := canvas.Get("dataset")
	dataset.Set("gameState", current.State.String())
	dataset.Set("score", strconv.Itoa(current.Score))
	dataset.Set("ball", strconv.Itoa(current.BallNumber))
	dataset.Set("ballsRemaining", strconv.Itoa(current.BallsRemaining))
	canvas.Call("setAttribute", "aria-label", fmt.Sprintf(
		"Neon Relay pinball: %s, score %d, ball %d of 3.",
		current.State, current.Score, min(current.BallNumber, 3),
	))
}
