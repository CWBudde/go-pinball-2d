package main

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"os"
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/game"
	"github.com/CWBudde/go-pinball-2d/internal/platform"
	"github.com/CWBudde/go-pinball-2d/internal/table"
	"github.com/gonutz/prototype/draw"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

func TestCachedResamplingPreservesTransparentCompositing(t *testing.T) {
	s, err := newSurface(30, 30)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	source := image.NewNRGBA(image.Rect(0, 0, 12, 12))
	for y := 0; y < 12; y++ {
		for x := 0; x < 12; x++ {
			source.SetNRGBA(x, y, color.NRGBA{R: 210, G: 150, B: 240, A: uint8((x + y) * 10)})
		}
	}
	s.images["fixture"] = source
	background := draw.RGB(.12, .2, .3)
	s.FillRect(0, 0, 30, 30, background)
	if err := s.DrawImageFileTo("fixture", 3, 4, 20, 18, 0); err != nil {
		t.Fatal(err)
	}
	first := append([]byte(nil), s.pixels.Pix...)
	s.FillRect(0, 0, 30, 30, background)
	if err := s.DrawImageFileTo("fixture", 3, 4, 20, 18, 0); err != nil {
		t.Fatal(err)
	}
	for i, v := range first {
		if v != s.pixels.Pix[i] {
			t.Fatal("cached draw differs from first draw")
		}
	}
	expected, err := newSurface(30, 30)
	if err != nil {
		t.Fatal(err)
	}
	defer expected.Close()
	expected.FillRect(0, 0, 30, 30, background)
	xdraw.CatmullRom.Transform(expected.pixels, f64.Aff3{20.0 / 12, 0, 3, 0, 18.0 / 12, 4}, source, source.Bounds(), xdraw.Over, nil)
	for i, v := range expected.pixels.Pix {
		if math.Abs(float64(v)-float64(first[i])) > 2 {
			t.Fatalf("cached composition differs at byte %d: %d vs %d", i, first[i], v)
		}
	}
}

func TestPreviewDrawDoesNotAdvanceEffects(t *testing.T) {
	s, err := newSurface(180, 270)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	// Tiny placeholder textures isolate response timing from asset decode cost.
	entries, err := os.ReadDir("../../assets/images")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		s.images["assets/images/"+entry.Name()] = image.NewRGBA(image.Rect(0, 0, 4, 4))
	}
	g := game.New(table.New(), nil)
	g.State = game.Playing
	preview := new(platform.PreviewRenderer)
	preview.Advance(g, 0, []game.Event{{Kind: game.BumperHit, ID: "bumper_left", At: g.Table.Bumpers[0].Center}})
	if err := preview.Draw(s, g); err != nil {
		t.Fatal(err)
	}
	first := append([]byte(nil), s.pixels.Pix...)
	if err := preview.Draw(s, g); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, s.pixels.Pix) {
		t.Fatal("drawing the same state advanced an effect")
	}
	preview.Advance(g, .12, nil)
	if err := preview.Draw(s, g); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(first, s.pixels.Pix) {
		t.Fatal("advancing effect time did not change the rendering")
	}
}

func TestSoftwareStrokeHasAntialiasedEdges(t *testing.T) {
	s, err := newSurface(20, 20)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.StrokeLine(2, 2, 17, 13, 3, draw.White)
	partial, solid := false, false
	for i := 3; i < len(s.pixels.Pix); i += 4 {
		a := s.pixels.Pix[i]
		partial = partial || (a > 0 && a < 255)
		solid = solid || a == 255
	}
	if !partial || !solid {
		t.Fatal("stroke lost solid core or antialiased edge coverage")
	}
}
