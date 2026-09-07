// Package table contains the data-driven definition of the Neon Relay
// playfield. Coordinates use the game's y-down logical coordinate system.
package table

import (
	"fmt"
	"math"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
)

const (
	// Width and Height are the fixed logical dimensions of the playfield.
	Width  = 720.0
	Height = 1080.0

	BallRadius = 14.0

	BumperScore     = 100
	SlingshotScore  = 50
	RolloverScore   = 250
	DropTargetScore = 500
)

// FeatureKind describes the rule-facing role of a named table feature. The
// zero value is intentionally not a valid kind, so missing catalog entries do
// not silently look like ordinary playfield geometry.
type FeatureKind string

const (
	FeatureWall       FeatureKind = "wall"
	FeatureBumper     FeatureKind = "bumper"
	FeatureSlingshot  FeatureKind = "slingshot"
	FeatureRollover   FeatureKind = "rollover"
	FeatureDropTarget FeatureKind = "drop_target"
	FeatureInlane     FeatureKind = "inlane"
	FeatureOutlane    FeatureKind = "outlane"
	FeaturePost       FeatureKind = "post"
	FeatureDrain      FeatureKind = "drain"
	FeatureFlipper    FeatureKind = "flipper"
	FeaturePlunger    FeatureKind = "plunger"
)

// Feature is the game-rule metadata for a collider, sensor, or mechanism ID.
// Score is sourced from the corresponding table definition rather than from a
// second set of rules in the game package.
type Feature struct {
	ID    string
	Kind  FeatureKind
	Score int
}

var (
	wallMaterial      = physics.Material{Restitution: 0.72, Friction: 0.16}
	postMaterial      = physics.Material{Restitution: 0.78, Friction: 0.12}
	bumperMaterial    = physics.Material{Restitution: 1.08, Friction: 0.04}
	slingshotMaterial = physics.Material{Restitution: 0.96, Friction: 0.08}
	targetMaterial    = physics.Material{Restitution: 0.44, Friction: 0.22}
)

// Bumper is both render metadata and a circular physics surface.
type Bumper struct {
	ID     string
	Center physics.Vec
	Radius float64
	Score  int
}

func (b Bumper) Collider() physics.CircleCollider {
	return physics.CircleCollider{ID: b.ID, Center: b.Center, Radius: b.Radius, Material: bumperMaterial}
}

// Slingshot describes a triangular visible and physical rubber. A game can
// score it directly from contacts carrying ID.
type Slingshot struct {
	ID       string
	Triangle [3]physics.Vec
	Radius   float64
	Score    int
}

func (s Slingshot) Colliders() [3]physics.LineCollider {
	var colliders [3]physics.LineCollider
	for i := range colliders {
		colliders[i] = physics.LineCollider{
			ID: s.ID,
			Segment: physics.Segment{
				A: s.Triangle[i],
				B: s.Triangle[(i+1)%len(s.Triangle)],
			},
			Radius:   s.Radius,
			Material: slingshotMaterial,
		}
	}
	return colliders
}

// Lane is a named scoring or routing sensor. The Segment and Radius also give
// the renderer enough information to draw its lamp or lane marking.
type Lane struct {
	ID      string
	Segment physics.Segment
	Radius  float64
	Score   int
}

func (l Lane) Sensor() physics.SegmentSensor {
	return physics.SegmentSensor{ID: l.ID, Segment: l.Segment, Radius: l.Radius}
}

// DropTarget is a thin physical target with a matching hit sensor. Games may
// rebuild the world without a target's collider after it has dropped.
type DropTarget struct {
	ID      string
	Segment physics.Segment
	Radius  float64
	Score   int
}

func (t DropTarget) Collider() physics.LineCollider {
	return physics.LineCollider{ID: t.ID, Segment: t.Segment, Radius: t.Radius, Material: targetMaterial}
}

func (t DropTarget) Sensor() physics.SegmentSensor {
	return physics.SegmentSensor{ID: t.ID, Segment: t.Segment, Radius: t.Radius + BallRadius*.25}
}

// Post is a circular static playfield post.
type Post struct {
	ID     string
	Center physics.Vec
	Radius float64
}

func (p Post) Collider() physics.CircleCollider {
	return physics.CircleCollider{ID: p.ID, Center: p.Center, Radius: p.Radius, Material: postMaterial}
}

// Drain is the ball-lost sensor spanning the open bottom of the table.
type Drain struct {
	ID       string
	Min, Max physics.Vec
}

func (d Drain) Sensor() physics.BoxSensor {
	return physics.BoxSensor{ID: d.ID, Min: d.Min, Max: d.Max}
}

// PlungerDefinition locates the physical plunger mechanism on the playfield.
type PlungerDefinition struct {
	ID        string
	Position  physics.Vec
	Mechanism physics.Plunger
}

// Definition is the complete immutable-by-convention description of one Neon
// Relay table. New returns fresh slices and fresh flipper mechanisms so separate
// games never share mutable state.
type Definition struct {
	OuterWalls  []physics.LineCollider
	ShooterLane []physics.LineCollider
	GuideWalls  []physics.LineCollider

	Bumpers       []Bumper
	Slingshots    []Slingshot
	RolloverLanes []Lane
	DropTargets   []DropTarget
	Inlanes       []Lane
	Outlanes      []Lane
	Posts         []Post

	Drain     Drain
	BallSpawn physics.Vec
	Plunger   PlungerDefinition
	Flippers  []*physics.Flipper

	// Features maps every collider, sensor, and mechanism ID to its rule-facing
	// metadata. New returns a fresh map for each definition.
	Features map[string]Feature
}

// New constructs the standard Neon Relay table.
func New() *Definition {
	line := func(id string, ax, ay, bx, by, radius float64) physics.LineCollider {
		return physics.LineCollider{
			ID: id, Segment: physics.Segment{A: physics.V(ax, ay), B: physics.V(bx, by)},
			Radius: radius, Material: wallMaterial,
		}
	}

	leftFlipper := physics.NewFlipper("flipper_left", physics.V(220, 925), 105, 16, .31, -.58)
	rightFlipper := physics.NewFlipper("flipper_right", physics.V(500, 925), 105, 16, math.Pi-.31, math.Pi+.58)

	d := &Definition{
		OuterWalls: []physics.LineCollider{
			line("wall_outer_left_upper", 40, 150, 55, 105, 5),
			line("wall_outer_top_left", 55, 105, 120, 55, 5),
			line("wall_outer_top", 120, 55, 565, 55, 5),
			line("wall_outer_left", 40, 150, 40, 815, 5),
			line("wall_outer_left_lower", 40, 815, 155, 1015, 5),
			line("wall_outer_left_drain", 155, 1015, 260, 1050, 5),
			line("wall_outer_right_lower", 625, 815, 510, 1015, 5),
			line("wall_outer_right_drain", 510, 1015, 510, 1050, 5),
		},
		ShooterLane: []physics.LineCollider{
			line("wall_shooter_outer_upper", 565, 55, 650, 85, 5),
			line("wall_shooter_outer_curve", 650, 85, 685, 145, 5),
			line("wall_shooter_outer", 685, 145, 685, 1045, 5),
			line("wall_shooter_inner_upper", 575, 108, 610, 165, 5),
			line("wall_shooter_inner_curve", 610, 165, 625, 230, 5),
			line("wall_shooter_inner", 625, 230, 625, 1045, 5),
			line("wall_shooter_bottom", 625, 1045, 685, 1045, 5),
		},
		GuideWalls: []physics.LineCollider{
			line("guide_left_outlane", 90, 730, 120, 810, 4),
			line("guide_left_inlane", 195, 823, 208, 882, 4),
			line("guide_left_flipper", 208, 882, 220, 901, 4),
			line("guide_right_outlane", 584, 730, 580, 810, 4),
			line("guide_right_inlane", 525, 823, 512, 882, 4),
			line("guide_right_flipper", 512, 882, 500, 901, 4),
		},
		Bumpers: []Bumper{
			{ID: "bumper_left", Center: physics.V(238, 320), Radius: 54, Score: BumperScore},
			{ID: "bumper_right", Center: physics.V(476, 320), Radius: 54, Score: BumperScore},
			{ID: "bumper_center", Center: physics.V(344, 627), Radius: 54, Score: BumperScore},
		},
		Slingshots: []Slingshot{
			{
				ID: "slingshot_left", Score: SlingshotScore, Radius: 9,
				Triangle: [3]physics.Vec{physics.V(185, 688), physics.V(267, 852), physics.V(195, 823)},
			},
			{
				ID: "slingshot_right", Score: SlingshotScore, Radius: 9,
				Triangle: [3]physics.Vec{physics.V(535, 688), physics.V(525, 823), physics.V(453, 852)},
			},
		},
		RolloverLanes: []Lane{
			{ID: "rollover_left", Segment: physics.Segment{A: physics.V(211, 185), B: physics.V(267, 185)}, Radius: 7, Score: RolloverScore},
			{ID: "rollover_center", Segment: physics.Segment{A: physics.V(331, 185), B: physics.V(387, 185)}, Radius: 7, Score: RolloverScore},
			{ID: "rollover_right", Segment: physics.Segment{A: physics.V(451, 185), B: physics.V(507, 185)}, Radius: 7, Score: RolloverScore},
		},
		DropTargets: []DropTarget{
			{ID: "target_relay_1", Segment: physics.Segment{A: physics.V(514, 414), B: physics.V(554, 434)}, Radius: 12, Score: DropTargetScore},
			{ID: "target_relay_2", Segment: physics.Segment{A: physics.V(501, 476), B: physics.V(541, 496)}, Radius: 12, Score: DropTargetScore},
			{ID: "target_relay_3", Segment: physics.Segment{A: physics.V(488, 538), B: physics.V(528, 558)}, Radius: 12, Score: DropTargetScore},
			{ID: "target_relay_4", Segment: physics.Segment{A: physics.V(475, 600), B: physics.V(515, 620)}, Radius: 12, Score: DropTargetScore},
		},
		Inlanes: []Lane{
			{ID: "inlane_left", Segment: physics.Segment{A: physics.V(155, 805), B: physics.V(195, 890)}, Radius: 12},
			{ID: "inlane_right", Segment: physics.Segment{A: physics.V(558, 805), B: physics.V(525, 890)}, Radius: 12},
		},
		Outlanes: []Lane{
			{ID: "outlane_left", Segment: physics.Segment{A: physics.V(70, 825), B: physics.V(135, 955)}, Radius: 12},
			{ID: "outlane_right", Segment: physics.Segment{A: physics.V(590, 825), B: physics.V(525, 955)}, Radius: 12},
		},
		Posts: []Post{
			{ID: "post_left_upper", Center: physics.V(105, 650), Radius: 13},
			{ID: "post_right_upper", Center: physics.V(585, 650), Radius: 13},
			{ID: "post_left_sling", Center: physics.V(185, 688), Radius: 14},
			{ID: "post_right_sling", Center: physics.V(535, 688), Radius: 14},
			{ID: "post_left_inlane", Center: physics.V(267, 852), Radius: 12},
			{ID: "post_right_inlane", Center: physics.V(453, 852), Radius: 12},
			{ID: "post_left_flipper", Center: physics.V(200, 900), Radius: 11},
			{ID: "post_right_flipper", Center: physics.V(520, 900), Radius: 11},
		},
		Drain:     Drain{ID: "drain", Min: physics.V(260, 1035), Max: physics.V(510, Height)},
		BallSpawn: physics.V(655, 985),
		Plunger: PlungerDefinition{
			ID: "plunger", Position: physics.V(655, 1020), Mechanism: physics.NewPlunger(physics.V(0, -1)),
		},
		Flippers: []*physics.Flipper{leftFlipper, rightFlipper},
	}

	// Upper guide shoulders frame three open-bottom rollover channels. Keep the
	// launch orbit above them clear; their top sits 77 units below the outer roof.
	for side, points := range [][]physics.Vec{
		quadratic(physics.V(136, 259), physics.V(98, 132), physics.V(210, 132), 12),
		quadratic(physics.V(572, 259), physics.V(600, 132), physics.V(508, 132), 12),
	} {
		for i := 1; i < len(points); i++ {
			a, b := points[i-1], points[i]
			d.GuideWalls = append(d.GuideWalls, line(fmt.Sprintf("guide_arch_%d_%02d", side, i), a.X, a.Y, b.X, b.Y, 7))
		}
	}
	for i, x := range []float64{188, 300, 420, 532} {
		d.GuideWalls = append(d.GuideWalls, line(fmt.Sprintf("guide_rollover_%d", i), x, 143, x, 218, 7))
	}

	d.Features = d.featureCatalog()
	return d
}

// Feature returns the rule-facing metadata for id.
func (d *Definition) Feature(id string) (Feature, bool) {
	if d == nil {
		return Feature{}, false
	}
	feature, ok := d.Features[id]
	return feature, ok
}

func (d *Definition) featureCatalog() map[string]Feature {
	features := make(map[string]Feature)
	add := func(id string, kind FeatureKind, score int) {
		if id == "" {
			panic("table: feature has an empty ID")
		}
		if _, exists := features[id]; exists {
			panic("table: duplicate feature ID " + id)
		}
		features[id] = Feature{ID: id, Kind: kind, Score: score}
	}
	for _, collider := range d.OuterWalls {
		add(collider.ID, FeatureWall, 0)
	}
	for _, collider := range d.ShooterLane {
		add(collider.ID, FeatureWall, 0)
	}
	for _, collider := range d.GuideWalls {
		add(collider.ID, FeatureWall, 0)
	}
	for _, bumper := range d.Bumpers {
		add(bumper.ID, FeatureBumper, bumper.Score)
	}
	for _, slingshot := range d.Slingshots {
		add(slingshot.ID, FeatureSlingshot, slingshot.Score)
	}
	for _, lane := range d.RolloverLanes {
		add(lane.ID, FeatureRollover, lane.Score)
	}
	for _, target := range d.DropTargets {
		add(target.ID, FeatureDropTarget, target.Score)
	}
	for _, lane := range d.Inlanes {
		add(lane.ID, FeatureInlane, lane.Score)
	}
	for _, lane := range d.Outlanes {
		add(lane.ID, FeatureOutlane, lane.Score)
	}
	for _, post := range d.Posts {
		add(post.ID, FeaturePost, 0)
	}
	add(d.Drain.ID, FeatureDrain, 0)
	for _, flipper := range d.Flippers {
		add(flipper.ID, FeatureFlipper, 0)
	}
	add(d.Plunger.ID, FeaturePlunger, 0)
	return features
}

// LineColliders returns all currently active static line surfaces. The returned
// slice is independent of the definition and is safe for a World to own.
func (d *Definition) LineColliders() []physics.LineCollider {
	if d == nil {
		return nil
	}
	count := len(d.OuterWalls) + len(d.ShooterLane) + len(d.GuideWalls) + 3*len(d.Slingshots) + len(d.DropTargets)
	lines := make([]physics.LineCollider, 0, count)
	lines = append(lines, d.OuterWalls...)
	lines = append(lines, d.ShooterLane...)
	lines = append(lines, d.GuideWalls...)
	for _, slingshot := range d.Slingshots {
		colliders := slingshot.Colliders()
		lines = append(lines, colliders[:]...)
	}
	for _, target := range d.DropTargets {
		lines = append(lines, target.Collider())
	}
	return lines
}

// CircleColliders returns all bumper and post surfaces.
func (d *Definition) CircleColliders() []physics.CircleCollider {
	if d == nil {
		return nil
	}
	circles := make([]physics.CircleCollider, 0, len(d.Bumpers)+len(d.Posts))
	for _, bumper := range d.Bumpers {
		circles = append(circles, bumper.Collider())
	}
	for _, post := range d.Posts {
		circles = append(circles, post.Collider())
	}
	return circles
}

// Sensors returns every rule-bearing non-contact trigger in stable playfield
// order. Inlanes and outlanes are visual routing markers, not game-rule
// sensors, so they are deliberately omitted.
func (d *Definition) Sensors() []physics.Sensor {
	if d == nil {
		return nil
	}
	count := len(d.RolloverLanes) + len(d.DropTargets) + 1
	sensors := make([]physics.Sensor, 0, count)
	for _, lane := range d.RolloverLanes {
		sensors = append(sensors, lane.Sensor())
	}
	for _, target := range d.DropTargets {
		sensors = append(sensors, target.Sensor())
	}
	sensors = append(sensors, d.Drain.Sensor())
	return sensors
}

// World builds a physics world backed by the table's current mechanisms.
func (d *Definition) World() physics.World {
	if d == nil {
		return physics.World{}
	}
	return physics.World{
		Gravity: physics.V(0, 760), Lines: d.LineColliders(), Circles: d.CircleColliders(),
		Flippers: d.Flippers, MaxIterations: 10, SafePosition: d.BallSpawn,
	}
}

// quadratic samples one authored guide curve for both collision and rendering.
func quadratic(a, control, b physics.Vec, steps int) []physics.Vec {
	points := make([]physics.Vec, steps+1)
	for i := range points {
		t := float64(i) / float64(steps)
		points[i] = a.Mul((1 - t) * (1 - t)).Add(control.Mul(2 * (1 - t) * t)).Add(b.Mul(t * t))
	}
	return points
}
