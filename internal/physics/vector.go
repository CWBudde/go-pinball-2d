package physics

import "math"

const geometryEpsilon = 1e-9

// Vec is a two-dimensional vector in logical playfield units.
type Vec struct {
	X float64
	Y float64
}

func V(x, y float64) Vec             { return Vec{X: x, Y: y} }
func (a Vec) Add(b Vec) Vec          { return Vec{a.X + b.X, a.Y + b.Y} }
func (a Vec) Sub(b Vec) Vec          { return Vec{a.X - b.X, a.Y - b.Y} }
func (a Vec) Mul(s float64) Vec      { return Vec{a.X * s, a.Y * s} }
func (a Vec) Div(s float64) Vec      { return Vec{a.X / s, a.Y / s} }
func (a Vec) Dot(b Vec) float64      { return a.X*b.X + a.Y*b.Y }
func (a Vec) LengthSquared() float64 { return a.Dot(a) }
func (a Vec) Length() float64        { return math.Sqrt(a.LengthSquared()) }
func (a Vec) Perp() Vec              { return Vec{-a.Y, a.X} }
func (a Vec) IsFinite() bool         { return finite(a.X) && finite(a.Y) }
func (a Vec) AlmostEqual(b Vec) bool {
	return a.Sub(b).LengthSquared() <= geometryEpsilon*geometryEpsilon
}

func (a Vec) Normalized() Vec {
	l := a.Length()
	if l <= geometryEpsilon || !finite(l) {
		return Vec{}
	}
	return a.Div(l)
}

func (a Vec) ClampLength(max float64) Vec {
	if max <= 0 {
		return Vec{}
	}
	lsq := a.LengthSquared()
	if !finite(lsq) {
		return Vec{}
	}
	if lsq <= max*max {
		return a
	}
	return a.Mul(max / math.Sqrt(lsq))
}

func Clamp(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
