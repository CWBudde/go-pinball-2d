package game

import "testing"

func TestStateString(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{Loading, "Loading"},
		{Attract, "Attract"},
		{BallReady, "Ball Ready"},
		{Playing, "Playing"},
		{Paused, "Paused"},
		{BallLost, "Ball Lost"},
		{GameOver, "Game Over"},
		{State(255), "State(255)"},
	}
	for _, test := range tests {
		if got := test.state.String(); got != test.want {
			t.Errorf("State(%d).String() = %q, want %q", test.state, got, test.want)
		}
	}
}
