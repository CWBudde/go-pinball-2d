package main

import "github.com/CWBudde/go-pinball-2d/internal/platform"

func main() {
	if err := platform.Run(); err != nil {
		panic(err)
	}
}
