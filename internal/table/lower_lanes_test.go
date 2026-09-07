package table

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
)

func TestLowerAssemblyCenteredOutsideShooterLane(t *testing.T) {
	d := New()
	var left, right float64
	for _, wall := range d.OuterWalls {
		if wall.ID == "wall_outer_left" {
			left = wall.Segment.A.X + wall.Radius
		}
	}
	for _, wall := range d.ShooterLane {
		if wall.ID == "wall_shooter_inner" {
			right = wall.Segment.A.X - wall.Radius
		}
	}
	center := (left + right) / 2
	if math.Abs((d.Flippers[0].Pivot.X+d.Flippers[1].Pivot.X)/2-center) > 1e-9 {
		t.Fatal("flippers are not centered within the playfield excluding the shooter lane")
	}
	if center >= Width/2 {
		t.Fatal("lower assembly must be left of cabinet center")
	}
	for _, lanes := range [][]Lane{d.Inlanes, d.Outlanes} {
		a := lanes[0].Segment.A.Add(lanes[0].Segment.B).Mul(.5)
		b := lanes[1].Segment.A.Add(lanes[1].Segment.B).Mul(.5)
		if math.Abs((a.X+b.X)/2-center) > 1e-9 || a.Y != b.Y {
			t.Fatal("lane pair is not mirrored about usable center")
		}
	}
}

// Probe continuous routes with a full-size ball, including the diagonal bends.
// Checking horizontal lane widths alone misses pinches against sloped rubbers.
func TestLowerLaneRoutesHaveBallClearance(t *testing.T) {
	routes := map[string][]physics.Vec{
		"inlane":  {physics.V(125, 700), physics.V(128, 780), physics.V(140, 820), physics.V(156, 847), physics.V(178, 867), physics.V(200, 877)},
		"outlane": {physics.V(68, 700), physics.V(68, 815), physics.V(105, 880), physics.V(160, 975), physics.V(240, 1010), physics.V(270, 1040)},
	}
	d := New()
	const margin = 2.0
	for name, path := range routes {
		for _, right := range []bool{false, true} {
			for i := 1; i < len(path); i++ {
				for j := 0; j <= 100; j++ {
					p := path[i-1].Add(path[i].Sub(path[i-1]).Mul(float64(j) / 100))
					if right {
						p.X = 2*PlayfieldCenter - p.X
					}
					check := func(id string, clearance float64) {
						t.Helper()
						if clearance < BallRadius+margin {
							t.Fatalf("%s right=%t at %+v: %s leaves radius %.2f, need %.2f", name, right, p, id, clearance, BallRadius+margin)
						}
					}
					for _, line := range d.LineColliders() {
						check(line.ID, physics.DistancePointSegment(p, line.Segment)-line.Radius)
					}
					for _, c := range d.CircleColliders() {
						check(c.ID, p.Sub(c.Center).Length()-c.Radius)
					}
					for _, f := range d.Flippers {
						check(f.ID, physics.DistancePointSegment(p, f.Segment())-f.Radius)
					}
				}
			}
		}
	}
}

func TestLowerLanesRouteBallsToFlipperOrDrain(t *testing.T) {
	for _, kind := range []string{"inlane", "outlane"} {
		for side := 0; side < 2; side++ {
			for _, offset := range []float64{-8, 0, 8} {
				for _, speed := range []float64{0, 180, 400} {
					for _, lateral := range []float64{-40, 0, 40} {
						t.Run(fmt.Sprintf("%s/side%d/x%g/v%g,%g", kind, side, offset, lateral, speed), func(t *testing.T) {
							d := New()
							world := d.World()
							lane := d.Inlanes[side]
							if kind == "outlane" {
								lane = d.Outlanes[side]
							}
							start := lane.Segment.A.Add(lane.Segment.B).Mul(.5)
							start.X += offset
							ball := physics.NewBall(start, BallRadius)
							ball.Velocity = physics.V(lateral, speed)
							for range 2400 {
								contacts := world.StepBall(&ball, 1.0/240)
								for _, contact := range contacts {
									if strings.HasPrefix(contact.ColliderID, "flipper_") {
										if kind == "outlane" {
											t.Fatalf("outlane returned to a flipper: %+v", ball.Position)
										}
										if contact.ColliderID != d.Flippers[side].ID {
											t.Fatalf("inlane fed wrong flipper: %s", contact.ColliderID)
										}
										return
									}
								}
								if d.Drain.Sensor().Overlaps(ball) {
									if kind == "inlane" {
										t.Fatal("inlane drained without touching its flipper")
									}
									return
								}
							}
							t.Fatalf("ball stuck or escaped: %+v, velocity %+v", ball.Position, ball.Velocity)
						})
					}
				}
			}
		}
	}
}
