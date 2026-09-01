//go:build js && wasm

package platform

import (
	"strconv"
	"syscall/js"
)

const highScoreKey = "neon-relay-high-score"

type browserStore struct{}

func (browserStore) LoadHighScore() int {
	return recoverJSValue(func() int {
		storage := js.Global().Get("localStorage")
		if !storage.Truthy() {
			return 0
		}
		value := storage.Call("getItem", highScoreKey)
		if value.IsNull() || value.IsUndefined() {
			return 0
		}
		score, err := strconv.Atoi(value.String())
		if err != nil || score < 0 {
			return 0
		}
		return score
	})
}

func (browserStore) SaveHighScore(score int) {
	_ = recoverJSValue(func() bool {
		storage := js.Global().Get("localStorage")
		if !storage.Truthy() {
			return false
		}
		storage.Call("setItem", highScoreKey, strconv.Itoa(score))
		return true
	})
}

func newHighScoreStore() browserStore { return browserStore{} }
