package physics

import "math"

// Sensor reports overlap only; it never changes ball motion. Game rules can
// assign lane, target, drain, or scoring meaning to its ID.
type Sensor interface {
	SensorID() string
	Overlaps(Ball) bool
}

type SegmentSensor struct {
	ID      string
	Segment Segment
	Radius  float64
}

func (s SegmentSensor) SensorID() string { return s.ID }
func (s SegmentSensor) Overlaps(b Ball) bool {
	return DistancePointSegment(b.Position, s.Segment) <= math.Max(0, s.Radius)+math.Max(0, b.Radius)
}

type BoxSensor struct {
	ID       string
	Min, Max Vec
}

func (s BoxSensor) SensorID() string { return s.ID }
func (s BoxSensor) Overlaps(b Ball) bool {
	closest := V(Clamp(b.Position.X, s.Min.X, s.Max.X), Clamp(b.Position.Y, s.Min.Y, s.Max.Y))
	return b.Position.Sub(closest).LengthSquared() <= b.Radius*b.Radius
}
