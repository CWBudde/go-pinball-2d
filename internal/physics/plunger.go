package physics

import "math"

type Plunger struct {
	Direction  Vec
	ChargeRate float64
	MinImpulse float64
	MaxImpulse float64
	Charge     float64
}

func NewPlunger(direction Vec) Plunger {
	return Plunger{
		Direction: direction.Normalized(), ChargeRate: 0.75,
		MinImpulse: 450, MaxImpulse: 1750,
	}
}

// Hold accumulates normalized charge in [0,1].
func (p *Plunger) Hold(dt float64) {
	if dt <= 0 || !finite(dt) {
		return
	}
	p.Charge = Clamp(p.Charge+math.Max(0, p.ChargeRate)*dt, 0, 1)
}

// Release consumes the stored charge and returns a launch velocity impulse.
func (p *Plunger) Release() Vec {
	charge := Clamp(p.Charge, 0, 1)
	p.Charge = 0
	if charge <= 0 {
		return Vec{}
	}
	strength := p.MinImpulse + (p.MaxImpulse-p.MinImpulse)*charge
	return p.Direction.Normalized().Mul(math.Max(0, strength))
}
