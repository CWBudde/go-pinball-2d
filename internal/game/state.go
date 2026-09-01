package game

import "fmt"

type State uint8

const (
	Loading State = iota
	Attract
	BallReady
	Playing
	Paused
	BallLost
	GameOver
)

func (s State) String() string {
	switch s {
	case Loading:
		return "Loading"
	case Attract:
		return "Attract"
	case BallReady:
		return "Ball Ready"
	case Playing:
		return "Playing"
	case Paused:
		return "Paused"
	case BallLost:
		return "Ball Lost"
	case GameOver:
		return "Game Over"
	default:
		return fmt.Sprintf("State(%d)", s)
	}
}

type Input struct {
	LeftFlipper  bool
	RightFlipper bool
	Plunger      bool
	StartPressed bool
	PausePressed bool
}
