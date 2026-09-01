package platform

import (
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
)

func TestViewportPreservesPortraitAspect(t *testing.T) {
	for _, screen := range [][2]int{{720, 1080}, {1920, 1080}, {360, 800}, {1, 1}} {
		view := newViewport(screen[0], screen[1])
		if view.width > max(1, screen[0]) || view.height > max(1, screen[1]) {
			t.Fatalf("viewport %dx%d exceeds screen %v", view.width, view.height, screen)
		}
		x, y := view.point(physics.V(720, 1080))
		if x != view.offsetX+view.width || y != view.offsetY+view.height {
			t.Fatalf("bottom-right mapped to %d,%d, want %d,%d", x, y, view.offsetX+view.width, view.offsetY+view.height)
		}
	}
}

func TestOddStrokeWidth(t *testing.T) {
	for _, test := range []struct {
		input int
		want  int
	}{
		{input: 0, want: 1},
		{input: 1, want: 1},
		{input: 2, want: 3},
		{input: 3, want: 3},
	} {
		if got := oddStrokeWidth(test.input); got != test.want {
			t.Errorf("oddStrokeWidth(%d) = %d, want %d", test.input, got, test.want)
		}
	}
}
