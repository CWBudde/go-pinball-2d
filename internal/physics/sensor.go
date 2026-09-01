package physics

import "math"

// Sensor reports overlap only; it never changes ball motion. Game rules can
// assign lane, target, drain, or scoring meaning to its ID.
type Sensor interface {
	SensorID() string
	Overlaps(Ball) bool
}

type CircleSensor struct {
	ID     string
	Center Vec
	Radius float64
}

func (s CircleSensor) SensorID() string { return s.ID }
func (s CircleSensor) Overlaps(b Ball) bool {
	r := math.Max(0, s.Radius) + math.Max(0, b.Radius)
	return b.Position.Sub(s.Center).LengthSquared() <= r*r
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

func OverlappingSensors(ball Ball, sensors []Sensor) []string {
	ids := make([]string, 0, 2)
	for _, sensor := range sensors {
		if sensor != nil && sensor.Overlaps(ball) {
			ids = append(ids, sensor.SensorID())
		}
	}
	return ids
}
