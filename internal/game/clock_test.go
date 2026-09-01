package game

import (
	"math"
	"testing"
)

func TestClockUsesFixedSteps(t *testing.T) {
	var clock Clock
	var got []float64
	steps := clock.Advance(1.0/60.0, func(dt float64) { got = append(got, dt) })
	if steps != 4 || len(got) != 4 {
		t.Fatalf("got %d fixed steps, want 4", steps)
	}
	for _, dt := range got {
		if dt != FixedStep {
			t.Fatalf("step is %g, want %g", dt, FixedStep)
		}
	}
}

func TestClockDropsSuspensionBacklog(t *testing.T) {
	var clock Clock
	steps := clock.Advance(30, func(float64) {})
	if steps != MaxSteps {
		t.Fatalf("got %d steps after suspension, want cap %d", steps, MaxSteps)
	}
	if clock.Alpha() >= 1 {
		t.Fatalf("backlog was not dropped: alpha=%g", clock.Alpha())
	}
}

func TestClockRejectsInvalidElapsedTime(t *testing.T) {
	for _, elapsed := range []float64{-1, math.NaN(), math.Inf(1)} {
		var clock Clock
		if got := clock.Advance(elapsed, func(float64) { t.Fatal("unexpected step") }); got != 0 {
			t.Fatalf("Advance(%v) = %d, want 0", elapsed, got)
		}
	}
}

func TestClockIsConsistentAcrossCommonRefreshRates(t *testing.T) {
	for _, refreshRate := range []int{60, 120, 144} {
		var clock Clock
		steps := 0
		for range refreshRate * 10 {
			steps += clock.Advance(1/float64(refreshRate), func(float64) {})
		}
		if steps != 10*240 {
			t.Errorf("%d Hz produced %d steps over 10 seconds, want %d", refreshRate, steps, 10*240)
		}
	}
}
