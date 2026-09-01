package physics

// Cooldowns suppress repeated events from persistent contacts. Time advances
// only when Advance is called, keeping behavior deterministic in fixed steps.
type Cooldowns struct {
	remaining map[string]float64
}

func (c *Cooldowns) Advance(dt float64) {
	if dt <= 0 || !finite(dt) {
		return
	}
	for key, left := range c.remaining {
		left -= dt
		if left <= 0 {
			delete(c.remaining, key)
		} else {
			c.remaining[key] = left
		}
	}
}

func (c *Cooldowns) Ready(key string) bool {
	return c.remaining == nil || c.remaining[key] <= 0
}

func (c *Cooldowns) Trigger(key string, duration float64) {
	if duration <= 0 || !finite(duration) {
		return
	}
	if c.remaining == nil {
		c.remaining = make(map[string]float64)
	}
	c.remaining[key] = duration
}

func (c *Cooldowns) Allow(key string, duration float64) bool {
	if !c.Ready(key) {
		return false
	}
	c.Trigger(key, duration)
	return true
}

func (c *Cooldowns) Clear() { clear(c.remaining) }
