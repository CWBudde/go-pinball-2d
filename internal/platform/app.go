// Package platform adapts the deterministic game to prototype/draw input,
// rendering, browser storage, and WAV playback.
package platform

import (
	"fmt"
	"time"

	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/table"
	"github.com/gonutz/prototype/draw"
)

const title = "Neon Relay"

type app struct {
	game          *game.Game
	renderer      renderer
	lastFrame     time.Time
	audioUnlocked bool
	iconSet       bool
	statusError   error
}

func Run() error {
	definition := table.New()
	application := &app{game: game.New(definition, newHighScoreStore())}
	return draw.RunWindow(title, int(table.Width), int(table.Height), application.update)
}

func (a *app) update(window draw.Window) {
	now := time.Now()
	elapsed := 0.0
	if !a.lastFrame.IsZero() {
		elapsed = now.Sub(a.lastFrame).Seconds()
	}
	a.lastFrame = now

	if window.WasKeyPressed(draw.KeyF) {
		window.SetFullscreen(!window.IsFullscreen())
	}

	input := game.Input{
		LeftFlipper:  window.IsKeyDown(draw.KeyA) || window.IsKeyDown(draw.KeyLeft),
		RightFlipper: window.IsKeyDown(draw.KeyD) || window.IsKeyDown(draw.KeyRight),
		Plunger:      window.IsKeyDown(draw.KeySpace) || window.IsKeyDown(draw.KeyDown),
		StartPressed: window.WasKeyPressed(draw.KeyEnter),
		PausePressed: window.WasKeyPressed(draw.KeyP),
	}
	if input.LeftFlipper || input.RightFlipper || input.Plunger || input.StartPressed || input.PausePressed {
		a.audioUnlocked = true
	}

	assetsReady := a.renderer.preload(window)
	if assetsReady {
		a.game.FinishLoading()
		if !a.iconSet {
			if err := window.SetIcon("assets/images/favicon.png"); err != nil {
				a.statusError = fmt.Errorf("favicon: %w", err)
			} else {
				a.iconSet = true
			}
		}
	}

	events := a.game.Update(elapsed, input)
	publishStatus(a.game)
	a.renderer.consume(events)
	if a.audioUnlocked {
		for _, event := range events {
			if path := soundFor(event.Kind); path != "" {
				if err := window.PlaySoundFile(path); err != nil {
					a.statusError = fmt.Errorf("audio %s: %w", path, err)
				}
			}
		}
	}
	a.renderer.draw(window, a.game, elapsed, a.statusError)
}

func soundFor(kind game.EventKind) string {
	switch kind {
	case game.FlipperFired:
		return "assets/audio/flipper.wav"
	case game.BumperHit, game.SlingshotHit:
		return "assets/audio/bumper.wav"
	case game.TargetDown, game.BankCompleted:
		return "assets/audio/target.wav"
	case game.BallLaunched:
		return "assets/audio/launch.wav"
	case game.JackpotAwarded:
		return "assets/audio/jackpot.wav"
	case game.BallDrained:
		return "assets/audio/drain.wav"
	case game.GameEnded:
		return "assets/audio/game-over.wav"
	default:
		return ""
	}
}
