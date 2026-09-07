// Command renderpreview saves frames from the real game engine to PNG, without
// opening a window or browser. See scripts/render-preview.sh for the portable
// Node/WebAssembly runner, which does not need native graphics libraries.
package main

import (
	"flag"
	"fmt"
	"image"
	imagedraw "image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/platform"
	"github.com/CWBudde/go-pinball-2d/internal/table"
)

func main() {
	out := flag.String("out", "output/preview", "directory for engine-rendered PNGs")
	width := flag.Int("width", 720, "frame width")
	height := flag.Int("height", 1080, "frame height")
	colliders := flag.Bool("colliders", false, "overlay actual contact geometry and sensors")
	blockout := flag.Bool("blockout", false, "render untextured layout and contact geometry")
	flag.Parse()
	if err := run(*out, *width, *height, *colliders, *blockout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(out string, width, height int, colliders, blockout bool) error {
	if width < 180 || height < 270 || width > 2880 || height > 4320 {
		return fmt.Errorf("frame dimensions must be within 180x270 and 2880x4320")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	window, err := newSurface(width, height)
	if err != nil {
		return err
	}
	defer window.Close()
	current := game.New(table.New(), nil)
	current.FinishLoading()
	capture := func(name string, want game.State) error {
		if current.State != want {
			return fmt.Errorf("%s: engine state %s, want %s", name, current.State, want)
		}
		if blockout {
			platform.RenderLayoutFrame(window, current)
		} else {
			if err := platform.RenderFrame(window, current); err != nil {
				return err
			}
			if colliders {
				platform.RenderCollisionOverlay(window, current)
			}
		}
		path := filepath.Join(out, name+".png")
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		encodeErr := png.Encode(file, window.pixels)
		closeErr := file.Close()
		if encodeErr != nil {
			return encodeErr
		}
		if closeErr != nil {
			return closeErr
		}
		fmt.Printf("%s: %s, score %d, ball (%.1f, %.1f)\n", path, current.State, current.Score, current.Ball.Position.X, current.Ball.Position.Y)
		return nil
	}
	step := func(frames int, input game.Input) {
		for range frames {
			current.Update(1.0/120, input)
		}
	}
	if err := capture("attract", game.Attract); err != nil {
		return err
	}
	step(1, game.Input{StartPressed: true})
	step(180, game.Input{Plunger: true})
	if err := capture("ready", game.BallReady); err != nil {
		return err
	}
	step(180, game.Input{})
	if err := capture("playing", game.Playing); err != nil {
		return err
	}
	left, right := current.World.Flippers[0].Angle, current.World.Flippers[1].Angle
	step(15, game.Input{LeftFlipper: true, RightFlipper: true})
	if current.World.Flippers[0].Angle == left || current.World.Flippers[1].Angle == right {
		return fmt.Errorf("flipper input did not move both mechanisms")
	}
	if err := capture("flippers", game.Playing); err != nil {
		return err
	}
	step(1, game.Input{PausePressed: true})
	if err := capture("paused", game.Paused); err != nil {
		return err
	}
	step(1, game.Input{})
	step(1, game.Input{PausePressed: true})
	if current.State != game.Playing {
		return fmt.Errorf("pause did not resume gameplay")
	}
	step(180, game.Input{})
	if current.Score == 0 {
		return fmt.Errorf("launch did not reach scoring features")
	}
	if err := capture("rally", game.Playing); err != nil {
		return err
	}
	return captureBumperStudy(out, window, current.Table)
}

// Crop the rendered rally pixels without rescaling or redrawing the sprites.
// Left is the Phase 2 study, right the original upper-right bumper for context.
func captureBumperStudy(out string, window *surface, definition *table.Definition) error {
	width, height := window.Size()
	scale := math.Min(float64(width)/table.Width, float64(height)/table.Height)
	offsetX := (width - int(math.Round(table.Width*scale))) / 2
	offsetY := (height - int(math.Round(table.Height*scale))) / 2
	size := int(math.Round(192 * scale))
	comparison := image.NewRGBA(image.Rect(0, 0, size*2, size))
	for i, id := range []string{table.MaterialStudyBumperID, "bumper_right"} {
		for _, bumper := range definition.Bumpers {
			if bumper.ID != id {
				continue
			}
			origin := image.Pt(offsetX+int(math.Round((bumper.Center.X-96)*scale)), offsetY+int(math.Round((bumper.Center.Y-96)*scale)))
			imagedraw.Draw(comparison, image.Rect(i*size, 0, (i+1)*size, size), window.pixels, origin, imagedraw.Src)
		}
	}
	path := filepath.Join(out, "bumper-comparison.png")
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	encodeErr := png.Encode(file, comparison)
	closeErr := file.Close()
	if encodeErr != nil {
		return encodeErr
	}
	if closeErr != nil {
		return closeErr
	}
	fmt.Printf("%s: material study (left), original treatment (right), %.1fx\n", path, scale)
	return nil
}
