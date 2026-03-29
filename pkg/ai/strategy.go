package ai

import "github.com/daghlny/scout_cli/pkg/engine"

// Strategy is the interface that all AI bots implement.
type Strategy interface {
	Name() string
	ChooseHandOrientation(hand *engine.Hand, state *engine.GameState) bool
	ChooseAction(state *engine.GameState, playerID int) engine.Action
}

// Candidate represents a potential combo found in a hand.
type Candidate struct {
	Start int
	Count int
	Combo *engine.Combo
}

// FindAllCombos enumerates all valid adjacent combos in a hand.
func FindAllCombos(hand *engine.Hand) []Candidate {
	var candidates []Candidate
	for start := 0; start < hand.Len(); start++ {
		for count := 1; count <= hand.Len()-start; count++ {
			vals, err := hand.AdjacentValues(start, count)
			if err != nil {
				continue
			}
			entries := hand.Entries()[start : start+count]
			cards := make([]engine.Card, count)
			for i, e := range entries {
				cards[i] = e.Card
			}
			combo, err := engine.ValidateCombo(vals, cards)
			if err != nil {
				continue
			}
			candidates = append(candidates, Candidate{
				Start: start,
				Count: count,
				Combo: combo,
			})
		}
	}
	return candidates
}

// CountAdjacentPairs counts how many adjacent pairs of cards share the same value
// or are consecutive. Used to evaluate hand quality for orientation choice.
func CountAdjacentPairs(hand *engine.Hand) int {
	vals := hand.ActiveValues()
	count := 0
	for i := 1; i < len(vals); i++ {
		diff := vals[i] - vals[i-1]
		if diff == 0 || diff == 1 || diff == -1 {
			count++
		}
	}
	return count
}
