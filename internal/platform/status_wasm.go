//go:build js && wasm

package platform

import (
	"fmt"
	"strconv"
	"syscall/js"

	"github.com/CWBudde/go-pinball-2d/internal/game"
)

var publishedStatus struct {
	canvas         js.Value
	state          game.State
	score          int
	ball           int
	ballsRemaining int
	initialized    bool
}

func publishStatus(current *game.Game) {
	if !publishedStatus.canvas.Truthy() {
		publishedStatus.canvas = js.Global().Get("document").Call("getElementById", "gameCanvas")
	}
	if !publishedStatus.canvas.Truthy() {
		return
	}
	dataset := publishedStatus.canvas.Get("dataset")
	stateChanged := !publishedStatus.initialized || publishedStatus.state != current.State
	scoreChanged := !publishedStatus.initialized || publishedStatus.score != current.Score
	ballChanged := !publishedStatus.initialized || publishedStatus.ball != current.BallNumber
	remainingChanged := !publishedStatus.initialized || publishedStatus.ballsRemaining != current.BallsRemaining
	if stateChanged {
		dataset.Set("gameState", current.State.String())
	}
	if scoreChanged {
		dataset.Set("score", strconv.Itoa(current.Score))
	}
	if ballChanged {
		dataset.Set("ball", strconv.Itoa(current.BallNumber))
	}
	if remainingChanged {
		dataset.Set("ballsRemaining", strconv.Itoa(current.BallsRemaining))
	}
	if stateChanged || scoreChanged || ballChanged {
		publishedStatus.canvas.Call("setAttribute", "aria-label", fmt.Sprintf(
			"Neon Relay pinball: %s, score %d, ball %d of 3.",
			current.State, current.Score, min(current.BallNumber, 3),
		))
	}
	publishedStatus.state = current.State
	publishedStatus.score = current.Score
	publishedStatus.ball = current.BallNumber
	publishedStatus.ballsRemaining = current.BallsRemaining
	publishedStatus.initialized = true
}
