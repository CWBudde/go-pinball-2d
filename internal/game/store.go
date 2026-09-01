package game

type HighScoreStore interface {
	LoadHighScore() int
	SaveHighScore(score int)
}

type MemoryStore struct {
	HighScore int
}

func (m *MemoryStore) LoadHighScore() int { return m.HighScore }
func (m *MemoryStore) SaveHighScore(score int) {
	if score > m.HighScore {
		m.HighScore = score
	}
}
