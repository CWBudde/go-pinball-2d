package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommittedAssetsAreFreshAndValid(t *testing.T) {
	files, err := generatedAssets()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(files), 28; got != want {
		t.Fatalf("generated %d files, want %d", got, want)
	}
	if err := checkAssets(filepath.Join("..", "..", "assets"), files); err != nil {
		t.Fatal(err)
	}

	dimensions := map[string]imageSize{
		"images/background.png":       {720, 1080},
		"images/table-shadows.png":    {720, 1080},
		"images/table-hardware.png":   {720, 1080},
		"images/table-foreground.png": {720, 1080},
		"images/logo.png":             {640, 200},
		"images/favicon.png":          {64, 64},
		"images/ball.png":             {64, 64},
		"images/flipper.png":          {180, 64},
		"images/bumper.png":           {128, 128},
		"images/bumper-material.png":  {384, 384},
		"images/bumper-patch.png":     {384, 384},
		"images/bumper-shadow.png":    {384, 384},
		"images/bumper-emission.png":  {384, 384},
		"images/post.png":             {48, 48},
		"images/target.png":           {64, 96},
		"images/target-down.png":      {64, 96},
		"images/lane-light.png":       {48, 96},
		"images/lane-light-off.png":   {48, 96},
		"images/plunger.png":          {56, 180},
		"images/glow.png":             {192, 192},
		"images/particle.png":         {32, 32},
	}
	for _, file := range files {
		if strings.HasSuffix(file.path, ".png") {
			config, err := png.DecodeConfig(bytes.NewReader(file.data))
			if err != nil {
				t.Errorf("decode %s: %v", file.path, err)
				continue
			}
			want := dimensions[file.path]
			if config.Width != want.width || config.Height != want.height {
				t.Errorf("%s dimensions = %dx%d, want %dx%d", file.path, config.Width, config.Height, want.width, want.height)
			}
		} else if strings.HasSuffix(file.path, ".wav") {
			validateWAV(t, file.path, file.data)
		}
	}
}

func TestPNGAssetComparisonUsesDecodedPixels(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 40, G: 50, B: 60, A: 128})
	encode := func(level png.CompressionLevel) []byte {
		t.Helper()
		var data bytes.Buffer
		if err := (&png.Encoder{CompressionLevel: level}).Encode(&data, img); err != nil {
			t.Fatal(err)
		}
		return data.Bytes()
	}
	fast := encode(png.NoCompression)
	compact := encode(png.BestCompression)
	if bytes.Equal(fast, compact) {
		t.Fatal("fixture encodings unexpectedly match")
	}
	equal, err := assetDataEqual("images/test.png", fast, compact)
	if err != nil || !equal {
		t.Fatalf("same decoded pixels = %t, error %v", equal, err)
	}
	img.SetNRGBA(1, 0, color.NRGBA{R: 41, G: 50, B: 60, A: 128})
	changed := encode(png.BestCompression)
	equal, err = assetDataEqual("images/test.png", changed, compact)
	if err != nil || equal {
		t.Fatalf("changed decoded pixels = %t, error %v", equal, err)
	}
}

func TestCheckAssetsRejectsOrphanedFiles(t *testing.T) {
	root := t.TempDir()
	files := []generatedFile{
		{path: "images/fixture.png", data: encodedFixturePNG(t)},
		{path: "audio/fixture.wav", data: []byte("sound")},
	}
	if err := writeAssets(root, files); err != nil {
		t.Fatal(err)
	}
	orphan := filepath.Join(root, "images", "old.png")
	if err := os.WriteFile(orphan, []byte("orphan"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := checkAssets(root, files)
	if err == nil || !strings.Contains(err.Error(), "orphaned") || !strings.Contains(err.Error(), orphan) {
		t.Fatalf("checkAssets() error = %v, want orphaned file", err)
	}
}

func encodedFixturePNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 1, G: 2, B: 3, A: 255})
	var data bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

type imageSize struct{ width, height int }

func validateWAV(t *testing.T, name string, data []byte) {
	t.Helper()
	if len(data) < 44 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" || string(data[12:16]) != "fmt " || string(data[36:40]) != "data" {
		t.Errorf("%s has an invalid canonical WAV header", name)
		return
	}
	u16 := func(offset int) uint16 { return binary.LittleEndian.Uint16(data[offset : offset+2]) }
	u32 := func(offset int) uint32 { return binary.LittleEndian.Uint32(data[offset : offset+4]) }
	if got := u32(4); got != uint32(len(data)-8) {
		t.Errorf("%s RIFF length = %d, want %d", name, got, len(data)-8)
	}
	if format, channels, rate, bits := u16(20), u16(22), u32(24), u16(34); format != 1 || channels != 1 || rate != sampleRate || bits != 16 {
		t.Errorf("%s format = PCM %d, channels %d, rate %d, bits %d; want PCM 1, mono, 44100 Hz, 16 bit", name, format, channels, rate, bits)
	}
	if got := u32(40); got != uint32(len(data)-44) {
		t.Errorf("%s data length = %d, want %d", name, got, len(data)-44)
	}
}
