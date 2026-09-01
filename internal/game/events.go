package game

import "github.com/CWBudde/go-pinball-2d/internal/physics"

type EventKind uint8

const (
	GameStarted EventKind = iota
	BallServed
	BallLaunched
	FlipperFired
	BumperHit
	SlingshotHit
	RolloverLit
	TargetDown
	BankCompleted
	JackpotAwarded
	BallDrained
	BonusAwarded
	GameEnded
)

type Event struct {
	Kind   EventKind
	ID     string
	Points int
	At     physics.Vec
}

type eventQueue struct {
	events []Event
}

func (q *eventQueue) emit(event Event) {
	q.events = append(q.events, event)
}

func (q *eventQueue) drain() []Event {
	if len(q.events) == 0 {
		return nil
	}
	result := append([]Event(nil), q.events...)
	q.events = q.events[:0]
	return result
}
