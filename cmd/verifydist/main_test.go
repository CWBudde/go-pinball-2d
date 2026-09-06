package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseReferences(t *testing.T) {
	index := []byte(`<link href="./favicon.png"><script src='./wasm_exec.js'></script><script>fetch("./main.wasm")</script>`)
	got := parseReferences(index)
	for _, path := range []string{"favicon.png", "wasm_exec.js", "main.wasm"} {
		if _, ok := got[path]; !ok {
			t.Fatalf("missing parsed reference %q in %#v", path, got)
		}
	}
}

func TestVerifyDistribution(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, repository, distribution string)
		wantError   string
		wantSuccess bool
	}{
		{name: "missing index", wantError: "read browser shell"},
		{
			name: "missing canvas",
			setup: func(t *testing.T, _, distribution string) {
				writeFixture(t, filepath.Join(distribution, "index.html"), "<html></html>")
			},
			wantError: "required gameCanvas",
		},
		{
			name: "missing nojekyll",
			setup: func(t *testing.T, repository, distribution string) {
				writeValidDistribution(t, repository, distribution, false)
			},
			wantError: "missing .nojekyll",
		},
		{
			name: "zero byte asset",
			setup: func(t *testing.T, repository, distribution string) {
				writeValidDistribution(t, repository, distribution, true)
				writeFixture(t, filepath.Join(distribution, "assets", "images", "ball.png"), "")
			},
			wantError: "empty or not a file",
		},
		{
			name: "success",
			setup: func(t *testing.T, repository, distribution string) {
				writeValidDistribution(t, repository, distribution, true)
			},
			wantSuccess: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := t.TempDir()
			distribution := filepath.Join(repository, "dist")
			if test.setup != nil {
				test.setup(t, repository, distribution)
			}
			err := verify(repository, distribution)
			if test.wantSuccess {
				if err != nil {
					t.Fatalf("verify() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("verify() error = %v, want text %q", err, test.wantError)
			}
		})
	}
}

func writeValidDistribution(t *testing.T, repository, distribution string, noJekyll bool) {
	t.Helper()
	writeFixture(t, filepath.Join(repository, "internal", "platform", "assets.go"), `package platform; const ball = "assets/images/ball.png"`)
	writeFixture(t, filepath.Join(distribution, "index.html"), `<canvas id="gameCanvas"></canvas><script src="./wasm_exec.js"></script><script>fetch("./main.wasm")</script>`)
	writeFixture(t, filepath.Join(distribution, "wasm_exec.js"), "wasm")
	writeFixture(t, filepath.Join(distribution, "main.wasm"), "module")
	writeFixture(t, filepath.Join(distribution, "assets", "images", "ball.png"), "pixels")
	if noJekyll {
		writeFixture(t, filepath.Join(distribution, ".nojekyll"), "")
	}
}

func writeFixture(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
