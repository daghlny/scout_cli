package engine

// ScoreRound calculates each player's score for the current round.
// roundEnderID is the player who triggered the round end (exempt from hand penalty).
func ScoreRound(players []PlayerState, roundEnderID int) []int {
	scores := make([]int, len(players))
	for i, p := range players {
		score := len(p.CollectedCards) + p.ScoutTokens
		if i != roundEnderID {
			score -= p.Hand.Len()
		}
		scores[i] = score
	}
	return scores
}
