package table

import (
	"math"
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
)

func TestRolloverChannelsAdmitAndReleaseBall(t *testing.T) {
	for _, lane := range New().RolloverLanes {
		t.Run(lane.ID, func(t *testing.T) {
			d := New()
			world := d.World()
			center := lane.Segment.A.Add(lane.Segment.B).Mul(.5)
			ball := physics.NewBall(physics.V(center.X, 105), BallRadius)
			ball.Velocity = physics.V(0, 220)
			crossed := false
			for range 720 {
				world.StepBall(&ball, 1.0/240)
				crossed = crossed || lane.Sensor().Overlaps(ball)
				if ball.Position.Y > 245 {
					if !crossed {
						t.Fatal("ball exited without crossing rollover sensor")
					}
					return
				}
			}
			t.Fatalf("ball stuck in rollover assembly: %+v", ball.Position)
		})
	}
}

func TestFlipperSweepClearsStaticHardware(t *testing.T) {
	d := New()
	for _, f := range d.Flippers {
		for i := 0; i <= 100; i++ {
			f.Angle = f.RestAngle + (f.ActiveAngle-f.RestAngle)*float64(i)/100
			for _, c := range d.CircleColliders() {
				if distance := physics.DistancePointSegment(c.Center, f.Segment()); distance < c.Radius+f.Radius-.01 {
					t.Fatalf("%s sweep overlaps %s at angle %.3f (distance %.3f)", f.ID, c.ID, f.Angle, distance)
				}
			}
			// Sample the complete blade to catch intersections, not only its endpoints.
			for j := 0; j <= 100; j++ {
				p := f.Pivot.Add(f.Tip().Sub(f.Pivot).Mul(float64(j) / 100))
				for _, line := range d.LineColliders() {
					if distance := physics.DistancePointSegment(p, line.Segment); distance < line.Radius+f.Radius-.01 {
						t.Fatalf("%s sweep overlaps %s at angle %.3f (distance %.3f)", f.ID, line.ID, f.Angle, distance)
					}
				}
			}
		}
	}
}

func TestSpriteAnchorsTrackPhysicalFlipper(t *testing.T) {
	for _, f := range New().Flippers {
		for _, angle := range []float64{f.RestAngle, (f.RestAngle + f.ActiveAngle) / 2, f.ActiveAngle} {
			f.Angle = angle
			scale := f.Length / FlipperFrame.Tip.Sub(FlipperFrame.Anchor).Length()
			placement := FlipperFrame.Place(f.Pivot, scale, angle)
			transform := func(source physics.Vec) physics.Vec {
				p := source.Sub(physics.V(FlipperFrame.Width/2, FlipperFrame.Height/2)).Mul(scale)
				sin, cos := math.Sincos(angle)
				return placement.Center.Add(physics.V(p.X*cos-p.Y*sin, p.X*sin+p.Y*cos))
			}
			if transform(FlipperFrame.Anchor).Sub(f.Pivot).Length() > 1e-9 || transform(FlipperFrame.Tip).Sub(f.Tip()).Length() > 1e-9 {
				t.Fatalf("%s visual pivot/tip detached at angle %v", f.ID, angle)
			}
		}
	}
}
