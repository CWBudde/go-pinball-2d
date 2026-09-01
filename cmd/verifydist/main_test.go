package main

import "testing"

func TestParseReferences(t *testing.T) {
	index := []byte(`<link href="./favicon.png"><script src='./wasm_exec.js'></script><script>fetch("./main.wasm")</script>`)
	got := parseReferences(index)
	for _, path := range []string{"favicon.png", "wasm_exec.js", "main.wasm"} {
		if _, ok := got[path]; !ok {
			t.Fatalf("missing parsed reference %q in %#v", path, got)
		}
	}
}
