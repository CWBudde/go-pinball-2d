package main

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommittedAssetsAreFreshAndValid(t *testing.T) {
	files, err := generatedAssets()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(files), 19; got != want {
		t.Fatalf("generated %d files, want %d", got, want)
	}
	if err := checkAssets(filepath.Join("..", "..", "assets"), files); err != nil {
		t.Fatal(err)
	}

	dimensions := map[string]imageSize{
		"images/background.png": {720, 1080},
		"images/logo.png":       {640, 200},
		"images/favicon.png":    {64, 64},
		"images/ball.png":       {64, 64},
		"images/flipper.png":    {180, 64},
		"images/bumper.png":     {128, 128},
		"images/post.png":       {48, 48},
		"images/target.png":     {64, 96},
		"images/lane-light.png": {48, 96},
		"images/plunger.png":    {56, 180},
		"images/glow.png":       {192, 192},
		"images/particle.png":   {32, 32},
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
