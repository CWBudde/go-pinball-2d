package main

import (
	"image"
	"image/color"
	imagedraw "image/draw"
	"math"
	"os"

	"github.com/gonutz/prototype/draw"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/f64"
	"golang.org/x/image/math/fixed"
)

// surface implements the drawing operations used by the real game renderer.
// The embedded interface intentionally panics if rendering attempts input,
// audio, or an unsupported drawing operation instead of silently omitting it.
type surface struct {
	draw.Window
	pixels   *image.RGBA
	images   map[string]image.Image
	typeface *opentype.Font
	faces    map[float32]font.Face
}

func newSurface(width, height int) (*surface, error) {
	typeface, err := opentype.Parse(gomono.TTF)
	if err != nil {
		return nil, err
	}
	return &surface{pixels: image.NewRGBA(image.Rect(0, 0, width, height)), images: make(map[string]image.Image), typeface: typeface, faces: make(map[float32]font.Face)}, nil
}

func (s *surface) Close() {
	for _, face := range s.faces {
		_ = face.Close()
	}
}

func (s *surface) Size() (int, int) { return s.pixels.Bounds().Dx(), s.pixels.Bounds().Dy() }

func rgba(c draw.Color) color.NRGBA {
	return color.NRGBA{R: uint8(c.R * 255), G: uint8(c.G * 255), B: uint8(c.B * 255), A: uint8(c.A * 255)}
}

func (s *surface) FillRect(x, y, width, height int, col draw.Color) {
	imagedraw.Draw(s.pixels, image.Rect(x, y, x+width, y+height), image.NewUniform(rgba(col)), image.Point{}, imagedraw.Over)
}

func (s *surface) DrawLine(ax, ay, bx, by int, col draw.Color) {
	dx, dy := bx-ax, by-ay
	steps := max(abs(dx), abs(dy))
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(max(1, steps))
		s.FillRect(ax+int(math.Round(float64(dx)*t)), ay+int(math.Round(float64(dy)*t)), 1, 1, col)
	}
}

func abs(x int) int { return max(x, -x) }

func (s *surface) DrawRect(x, y, width, height int, col draw.Color) {
	s.DrawLine(x, y, x+width-1, y, col)
	s.DrawLine(x, y+height-1, x+width-1, y+height-1, col)
	s.DrawLine(x, y, x, y+height-1, col)
	s.DrawLine(x+width-1, y, x+width-1, y+height-1, col)
}

func (s *surface) DrawEllipse(x, y, width, height int, col draw.Color) {
	if width <= 0 || height <= 0 {
		return
	}
	cx, cy := float64(x)+float64(width-1)/2, float64(y)+float64(height-1)/2
	rx, ry := float64(width-1)/2, float64(height-1)/2
	steps := max(12, int(math.Ceil(2*math.Pi*math.Max(rx, ry))))
	for i := 0; i < steps; i++ {
		a, b := float64(i)*2*math.Pi/float64(steps), float64(i+1)*2*math.Pi/float64(steps)
		s.DrawLine(int(math.Round(cx+rx*math.Cos(a))), int(math.Round(cy+ry*math.Sin(a))), int(math.Round(cx+rx*math.Cos(b))), int(math.Round(cy+ry*math.Sin(b))), col)
	}
}

func (s *surface) FillEllipse(x, y, width, height int, col draw.Color) {
	if width <= 0 || height <= 0 {
		return
	}
	rx, ry := float64(width)/2, float64(height)/2
	cx, cy := float64(x)+rx, float64(y)+ry
	for row := max(y, s.pixels.Bounds().Min.Y); row < min(y+height, s.pixels.Bounds().Max.Y); row++ {
		t := (float64(row) + .5 - cy) / ry
		extent := rx * math.Sqrt(max(0, 1-t*t))
		left, right := int(math.Ceil(cx-extent-.5)), int(math.Floor(cx+extent-.5))
		if right >= left {
			s.FillRect(left, row, right-left+1, 1, col)
		}
	}
}

func (s *surface) load(path string) (image.Image, error) {
	if img, ok := s.images[path]; ok {
		return img, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err == nil {
		s.images[path] = img
	}
	return img, err
}

func (s *surface) ImageSize(path string) (int, int, error) {
	img, err := s.load(path)
	if err != nil {
		return 0, 0, err
	}
	return img.Bounds().Dx(), img.Bounds().Dy(), nil
}

func (s *surface) DrawImageFileTo(path string, x, y, width, height, rotation int) error {
	img, err := s.load(path)
	if err != nil {
		return err
	}
	a := float64(rotation) * math.Pi / 180
	cos, sin := math.Cos(a), math.Sin(a)
	sx, sy := float64(width)/float64(img.Bounds().Dx()), float64(height)/float64(img.Bounds().Dy())
	cx, cy := float64(x)+float64(width)/2, float64(y)+float64(height)/2
	transform := f64.Aff3{cos * sx, -sin * sy, cx - cos*float64(width)/2 + sin*float64(height)/2, sin * sx, cos * sy, cy - sin*float64(width)/2 - cos*float64(height)/2}
	xdraw.CatmullRom.Transform(s.pixels, transform, img, img.Bounds(), xdraw.Over, nil)
	return nil
}

func (s *surface) face(scale float32) font.Face {
	if face, ok := s.faces[scale]; ok {
		return face
	}
	face, err := opentype.NewFace(s.typeface, &opentype.FaceOptions{Size: 24 * float64(scale), DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		panic(err)
	}
	s.faces[scale] = face
	return face
}

func (s *surface) GetScaledTextSize(text string, scale float32) (int, int) {
	face := s.face(scale)
	return font.MeasureString(face, text).Ceil(), face.Metrics().Height.Ceil()
}

func (s *surface) DrawScaledText(text string, x, y int, scale float32, col draw.Color) {
	face := s.face(scale)
	drawer := font.Drawer{Dst: s.pixels, Src: image.NewUniform(rgba(col)), Face: face, Dot: fixed.P(x, y+face.Metrics().Ascent.Ceil())}
	drawer.DrawString(text)
}
