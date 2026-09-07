// Package display defines the original instrument lettering shared by the
// asset generator and every renderer. Coordinates are in a four-by-six grid.
package display

import "strings"

const (
	Characters = " ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789:/%+.-?"
	Columns    = 16
	Rows       = 3
	CellWidth  = 32
	CellHeight = 48
)

const (
	Cyan = iota
	Silver
	Amber
	Pink
	Dim
	Colors
)

type (
	Point = [2]float64
	Path  = []Point
)

// Glyph paths have open counters and clipped corners that survive minification.
var Glyphs = map[rune][]Path{
	'A': {{{0, 6}, {0, 1}, {1, 0}, {3, 0}, {4, 1}, {4, 6}}, {{0, 3}, {4, 3}}},
	'B': {{{0, 6}, {0, 0}, {3, 0}, {4, 1}, {4, 2}, {3, 3}, {0, 3}}, {{3, 3}, {4, 4}, {4, 5}, {3, 6}, {0, 6}}},
	'C': {{{4, 0}, {1, 0}, {0, 1}, {0, 5}, {1, 6}, {4, 6}}},
	'D': {{{0, 6}, {0, 0}, {3, 0}, {4, 1}, {4, 5}, {3, 6}, {0, 6}}},
	'E': {{{4, 0}, {0, 0}, {0, 6}, {4, 6}}, {{0, 3}, {3.4, 3}}},
	'F': {{{0, 6}, {0, 0}, {4, 0}}, {{0, 3}, {3.4, 3}}},
	'G': {{{4, 0}, {1, 0}, {0, 1}, {0, 5}, {1, 6}, {4, 6}, {4, 3}, {2, 3}}},
	'H': {{{0, 0}, {0, 6}}, {{4, 0}, {4, 6}}, {{0, 3}, {4, 3}}},
	'I': {{{1, 0}, {3, 0}}, {{2, 0}, {2, 6}}, {{1, 6}, {3, 6}}},
	'J': {{{0, 5}, {1, 6}, {3, 6}, {4, 5}, {4, 0}}},
	'K': {{{0, 0}, {0, 6}}, {{4, 0}, {0, 3}, {4, 6}}},
	'L': {{{0, 0}, {0, 6}, {4, 6}}},
	'M': {{{0, 6}, {0, 0}, {2, 3}, {4, 0}, {4, 6}}},
	'N': {{{0, 6}, {0, 0}, {4, 6}, {4, 0}}},
	'O': {{{1, 0}, {3, 0}, {4, 1}, {4, 5}, {3, 6}, {1, 6}, {0, 5}, {0, 1}, {1, 0}}},
	'P': {{{0, 6}, {0, 0}, {3, 0}, {4, 1}, {4, 2}, {3, 3}, {0, 3}}},
	'Q': {{{1, 0}, {3, 0}, {4, 1}, {4, 5}, {3, 6}, {1, 6}, {0, 5}, {0, 1}, {1, 0}}, {{2, 4}, {4, 6}}},
	'R': {{{0, 6}, {0, 0}, {3, 0}, {4, 1}, {4, 2}, {3, 3}, {0, 3}}, {{2, 3}, {4, 6}}},
	'S': {{{4, 0}, {1, 0}, {0, 1}, {0, 2}, {1, 3}, {3, 3}, {4, 4}, {4, 5}, {3, 6}, {0, 6}}},
	'T': {{{0, 0}, {4, 0}}, {{2, 0}, {2, 6}}},
	'U': {{{0, 0}, {0, 5}, {1, 6}, {3, 6}, {4, 5}, {4, 0}}},
	'V': {{{0, 0}, {0, 3}, {2, 6}, {4, 3}, {4, 0}}},
	'W': {{{0, 0}, {0, 6}, {2, 3}, {4, 6}, {4, 0}}},
	'X': {{{0, 0}, {4, 6}}, {{4, 0}, {0, 6}}},
	'Y': {{{0, 0}, {2, 3}, {4, 0}}, {{2, 3}, {2, 6}}},
	'Z': {{{0, 0}, {4, 0}, {0, 6}, {4, 6}}},
	':': {{{2, 2}, {2, 2.2}}, {{2, 4.6}, {2, 4.8}}},
	'/': {{{0, 6}, {4, 0}}},
	'%': {{{0, 6}, {4, 0}}, {{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}, {{3, 5}, {4, 5}, {4, 6}, {3, 6}, {3, 5}}},
	'+': {{{0, 3}, {4, 3}}, {{2, 1}, {2, 5}}},
	'.': {{{2, 5.8}, {2, 6}}},
	'-': {{{0, 3}, {4, 3}}},
	'?': {{{0, 1}, {1, 0}, {3, 0}, {4, 1}, {4, 2}, {2, 3}, {2, 4}}, {{2, 5.8}, {2, 6}}},
}

// Numerals use seven isolated segments, like the table's inset LED instruments.
func Paths(ch rune) []Path {
	if ch < '0' || ch > '9' {
		return Glyphs[ch]
	}
	masks := [...]uint8{0x3f, 0x06, 0x5b, 0x4f, 0x66, 0x6d, 0x7d, 0x07, 0x7f, 0x6f}
	segments := [...]Path{
		{{.4, 0}, {3.6, 0}},
		{{4, .4}, {4, 2.6}},
		{{4, 3.4}, {4, 5.6}},
		{{.4, 6}, {3.6, 6}},
		{{0, 3.4}, {0, 5.6}},
		{{0, .4}, {0, 2.6}},
		{{.4, 3}, {3.6, 3}},
	}
	var paths []Path
	for i, path := range segments {
		if masks[ch-'0']&(1<<i) != 0 {
			paths = append(paths, path)
		}
	}
	return paths
}

func Cell(ch rune, tone int) (x, y int) {
	i := strings.IndexRune(Characters, ch)
	if i < 0 {
		i = strings.IndexRune(Characters, '?')
	}
	return i % Columns * CellWidth, (i/Columns + tone*Rows) * CellHeight
}

func Width(text string, height float64) float64 {
	return float64(len([]rune(text))) * height * CellWidth / CellHeight
}
