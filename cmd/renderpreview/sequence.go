package main

import (
	"fmt"
	"image"
	imagedraw "image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/platform"
	"github.com/CWBudde/go-pinball-2d/internal/table"
	"github.com/gonutz/prototype/draw"
)

// These fixtures position the real ball immediately before a contact. They
// require the actual game event, then park the ball in clear space to isolate
// recovery from additional hits. No synthetic events are injected.
func runSequence(out string, width, height int, selected string, colliders bool) error {
	if width < 180 || height < 270 || width > 2880 || height > 4320 {
		return fmt.Errorf("invalid capture dimensions")
	}
	names := []string{"bumper", "sling", "lane", "bank", "plunger"}
	if selected != "all" {
		found := false
		for _, name := range names {
			if name == selected {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("unknown sequence %q", selected)
		}
		names = []string{selected}
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	window, err := newSurface(width, height)
	if err != nil {
		return err
	}
	defer window.Close()
	for _, name := range names {
		if err := captureSequence(out, name, window, colliders); err != nil {
			return err
		}
	}
	return nil
}

func captureSequence(out, name string, window *surface, colliders bool) error {
	current := game.New(table.New(), nil)
	current.FinishLoading()
	preview := new(platform.PreviewRenderer)
	step := func(input game.Input) []game.Event {
		events := current.Update(game.FixedStep, input)
		preview.Advance(current, game.FixedStep, events)
		return events
	}
	step(game.Input{StartPressed: true})
	bounds := map[string]image.Rectangle{
		"bumper":  image.Rect(145, 225, 332, 420),
		"sling":   image.Rect(125, 687, 275, 901),
		"lane":    image.Rect(202, 128, 279, 239),
		"bank":    image.Rect(446, 386, 582, 651),
		"plunger": image.Rect(630, 968, 680, 1042),
	}[name]
	width, height := window.Size()
	scale := math.Min(float64(width)/table.Width, float64(height)/table.Height)
	offset := image.Pt((width-int(table.Width*scale))/2, (height-int(table.Height*scale))/2)
	region := image.Rect(int(float64(bounds.Min.X)*scale), int(float64(bounds.Min.Y)*scale), int(float64(bounds.Max.X)*scale), int(float64(bounds.Max.Y)*scale)).Add(offset)
	var frames []image.Image
	var labels []string
	capture := func(label string) error {
		if err := preview.Draw(window, current); err != nil {
			return err
		}
		if colliders {
			platform.RenderCollisionOverlay(window, current)
		}
		path := filepath.Join(out, name+"-"+label+".png")
		if err := savePNG(path, window.pixels); err != nil {
			return err
		}
		crop := image.NewRGBA(image.Rect(0, 0, region.Dx(), region.Dy()))
		imagedraw.Draw(crop, crop.Bounds(), window.pixels, region.Min, imagedraw.Src)
		frames = append(frames, crop)
		labels = append(labels, label)
		fmt.Printf("%s: %s, score %d\n", path, current.State, current.Score)
		return nil
	}
	if name == "plunger" {
		if err := capture("idle"); err != nil {
			return err
		}
		for range 180 {
			step(game.Input{Plunger: true})
		}
		if err := capture("half"); err != nil {
			return err
		}
		for range 180 {
			step(game.Input{Plunger: true})
		}
		if err := capture("charged"); err != nil {
			return err
		}
		launched := false
		for _, e := range step(game.Input{}) {
			launched = launched || e.Kind == game.BallLaunched
		}
		if !launched {
			return fmt.Errorf("plunger fixture did not launch")
		}
		if err := capture("released"); err != nil {
			return err
		}
	} else {
		for range 360 {
			step(game.Input{Plunger: true})
		}
		step(game.Input{})
		seedContact(current, name, 0)
		if err := capture("idle"); err != nil {
			return err
		}
		count := 1
		if name == "bank" {
			count = 4
		}
		for index := 0; index < count; index++ {
			wanted := seedContact(current, name, index)
			hit := false
			for attempt := 0; attempt < 240 && !hit; attempt++ {
				for _, event := range step(game.Input{}) {
					if event.Kind == wanted {
						hit = true
						fmt.Printf("%s: actual event %d on %s at (%.1f, %.1f)\n", name, event.Kind, event.ID, event.At.X, event.At.Y)
					}
				}
			}
			if !hit {
				return fmt.Errorf("%s fixture never produced event %d", name, wanted)
			}
		}
		if err := capture("impact"); err != nil {
			return err
		}
		// Isolate recovery while keeping the real bank reset timer running.
		current.Ball.Position = physics.V(table.PlayfieldCenter, 770)
		current.Ball.Velocity = physics.Vec{}
		current.World.Gravity = physics.Vec{}
		times := []struct {
			label   string
			seconds float64
		}{{"settle", .05}, {"recovered", .6}}
		if name == "bank" {
			times = []struct {
				label   string
				seconds float64
			}{{"chase", .28}, {"rise", 1.06}, {"recovered", 1.25}}
		}
		steps := 0
		for _, frame := range times {
			target := int(math.Round(frame.seconds / game.FixedStep))
			for steps < target {
				step(game.Input{})
				steps++
			}
			if err := capture(frame.label); err != nil {
				return err
			}
		}
	}
	return saveSequenceSheet(filepath.Join(out, name+"-sequence.png"), frames, labels, scale)
}

func seedContact(current *game.Game, name string, index int) game.EventKind {
	var center, normal physics.Vec
	var radius float64
	kind := game.BumperHit
	switch name {
	case "bumper":
		b := current.Table.Bumpers[0]
		center = b.Center
		normal = physics.V(0, -1)
		radius = b.Radius
	case "sling":
		s := current.Table.Slingshots[0]
		center = s.Triangle[0].Add(s.Triangle[1]).Mul(.5)
		delta := s.Triangle[1].Sub(s.Triangle[0])
		normal = physics.V(delta.Y, -delta.X).Normalized()
		radius = s.Radius
		kind = game.SlingshotHit
	case "lane":
		lane := current.Table.RolloverLanes[0]
		center = lane.Segment.A.Add(lane.Segment.B).Mul(.5)
		normal = physics.V(0, -1)
		radius = lane.Radius
		kind = game.RolloverLit
	case "bank":
		t := current.Table.DropTargets[index]
		center = t.Segment.A.Add(t.Segment.B).Mul(.5)
		delta := t.Segment.B.Sub(t.Segment.A)
		normal = physics.V(delta.Y, -delta.X).Normalized()
		radius = t.Sensor().Radius
		kind = game.TargetDown
	}
	current.Ball = physics.NewBall(center.Add(normal.Mul(radius+table.BallRadius+6)), table.BallRadius)
	current.Ball.Velocity = normal.Mul(-180)
	return kind
}

func savePNG(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	encodeErr := png.Encode(file, img)
	closeErr := file.Close()
	if encodeErr != nil {
		return encodeErr
	}
	return closeErr
}

func saveSequenceSheet(path string, frames []image.Image, labels []string, scale float64) error {
	gap, header := int(8*scale), int(28*scale)
	w, h := frames[0].Bounds().Dx(), frames[0].Bounds().Dy()
	sheet, err := newSurface((w+gap)*len(frames)+gap, h+header+gap)
	if err != nil {
		return err
	}
	defer sheet.Close()
	width, height := sheet.Size()
	sheet.FillRect(0, 0, width, height, draw.RGB(.03, .045, .05))
	for i, frame := range frames {
		x := gap + i*(w+gap)
		imagedraw.Draw(sheet.pixels, image.Rect(x, header, x+w, header+h), frame, image.Point{}, imagedraw.Src)
		sheet.DrawScaledText(labels[i], x, int(5*scale), float32(scale*.45), draw.RGB(.5, .85, .9))
	}
	if err := savePNG(path, sheet.pixels); err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}
