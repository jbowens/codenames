package codenames

import "time"

type gameCompletedEvent struct {
	ID           string
	CreatedAt    time.Time
	CompletedAt  time.Time
	RoundsPlayed int
}

func (s *Server) gameOver(g *Game) {
	if s.Events == nil {
		return
	}

	s.Events.Log("game_completed", gameCompletedEvent{
		ID:           g.ID,
		CreatedAt:    g.CreatedAt,
		CompletedAt:  time.Now(),
		RoundsPlayed: g.Round,
	})
}
