package physics_test

import (
	"testing"

	"github.com/CWBudde/go-pinball-2d/internal/physics"
	"github.com/CWBudde/go-pinball-2d/internal/table"
)

func BenchmarkStepBall(b *testing.B) {
	definition := table.New()
	world := definition.World()
	ball := physics.NewBall(physics.V(360, 500), table.BallRadius)
	ball.Velocity = physics.V(900, 700)
	b.ReportAllocs()
	for b.Loop() {
		stepBall := ball
		world.StepBall(&stepBall, 1.0/240.0)
	}
}
