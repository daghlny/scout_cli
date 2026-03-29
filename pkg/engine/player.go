package engine

// PlayerState holds all mutable state for a player within a game.
type PlayerState struct {
	ID             int
	Name           string
	Hand           *Hand
	CollectedCards []Card // cards won by beating table combos
	ScoutTokens    int    // earned when opponents scout your combo
	UsedScoutShow  bool   // true if Scout&Show chip has been used this round
	RoundScores    []int  // accumulated score per completed round
}

// TotalScore returns the sum of all round scores.
func (p *PlayerState) TotalScore() int {
	total := 0
	for _, s := range p.RoundScores {
		total += s
	}
	return total
}

// ResetForRound resets per-round state (hand is set separately during deal).
func (p *PlayerState) ResetForRound() {
	p.CollectedCards = nil
	p.ScoutTokens = 0
	p.UsedScoutShow = false
}
