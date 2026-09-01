// Package game owns deterministic game rules and fixed-step simulation timing.
package game

import "math"

const (
	// FixedStep is deliberately smaller than a rendered frame so fast pinballs
	// cannot cross thin table features between collision checks.
	FixedStep = 1.0 / 240.0
	MaxFrame  = 0.125
	MaxSteps  = 30
)

// Clock converts irregular rendered frames into deterministic simulation steps.
// Any excess after MaxSteps is discarded; catching up after a suspended browser
// tab is less useful than keeping the solver stable.
type Clock struct {
	accumulator float64
}

func (c *Clock) Advance(elapsed float64, simulate func(dt float64)) int {
	if elapsed < 0 || math.IsNaN(elapsed) || math.IsInf(elapsed, 0) {
		elapsed = 0
	}
	if elapsed > MaxFrame {
		elapsed = MaxFrame
	}
	c.accumulator += elapsed

	steps := 0
	for c.accumulator >= FixedStep && steps < MaxSteps {
		simulate(FixedStep)
		c.accumulator -= FixedStep
		steps++
	}
	if steps == MaxSteps && c.accumulator >= FixedStep {
		c.accumulator = math.Mod(c.accumulator, FixedStep)
	}
	return steps
}
