package ai

import (
	"math/rand"

	"github.com/daghlny/scout_cli/pkg/engine"
)

// GreedyBot always tries to Show the strongest possible combo.
// If it can't Show, it Scouts the highest-value card from the table.
type GreedyBot struct {
	rng *rand.Rand
}

func NewGreedyBot(rng *rand.Rand) *GreedyBot {
	return &GreedyBot{rng: rng}
}

func (b *GreedyBot) Name() string { return "Greedy" }

func (b *GreedyBot) ChooseHandOrientation(hand *engine.Hand, state *engine.GameState) bool {
	// Try both orientations, pick the one with more adjacent pairs
	original := CountAdjacentPairs(hand)
	hand.FlipAll()
	flipped := CountAdjacentPairs(hand)
	hand.FlipAll() // restore
	return flipped > original
}

func (b *GreedyBot) ChooseAction(state *engine.GameState, playerID int) engine.Action {
	actions := engine.ValidActions(state)

	// Separate by type
	var shows, scouts, scoutShows []engine.Action
	for _, a := range actions {
		switch a.Type {
		case engine.ActionShow:
			shows = append(shows, a)
		case engine.ActionScout:
			scouts = append(scouts, a)
		case engine.ActionScoutAndShow:
			scoutShows = append(scoutShows, a)
		}
	}

	// Prefer Show: pick the one with most cards, then highest min value
	if len(shows) > 0 {
		return b.bestShow(shows, state)
	}

	// Try Scout & Show if available
	if len(scoutShows) > 0 {
		return b.bestShow(scoutShows, state)
	}

	// Fall back to Scout: pick the end with the higher value
	if len(scouts) > 0 {
		return b.bestScout(scouts, state)
	}

	// Should never reach here if ValidActions is correct
	return actions[0]
}

func (b *GreedyBot) bestShow(actions []engine.Action, state *engine.GameState) engine.Action {
	hand := state.Players[state.CurrentPlayer].Hand
	best := actions[0]
	bestSize := 0
	bestMin := 0

	for _, a := range actions {
		start := a.ShowStart
		count := a.ShowCount

		// For ScoutAndShow, we need to simulate the scout first
		h := hand
		if a.Type == engine.ActionScoutAndShow {
			simHand := engine.NewHand(hand.Entries())
			entry := getScoutEntry(state, a.ScoutFromLeft, a.ScoutFlip)
			_ = simHand.Insert(entry, a.ScoutInsertAt)
			h = simHand
			start = a.ShowStart
			count = a.ShowCount
		}

		vals, err := h.AdjacentValues(start, count)
		if err != nil {
			continue
		}
		entries := h.Entries()[start : start+count]
		cards := make([]engine.Card, count)
		for i, e := range entries {
			cards[i] = e.Card
		}
		combo, err := engine.ValidateCombo(vals, cards)
		if err != nil {
			continue
		}

		size := combo.Size()
		min := combo.MinValue()
		if size > bestSize || (size == bestSize && min > bestMin) {
			best = a
			bestSize = size
			bestMin = min
		}
	}
	return best
}

func (b *GreedyBot) bestScout(actions []engine.Action, state *engine.GameState) engine.Action {
	// Pick the scout that gets the higher-value card
	best := actions[0]
	bestVal := 0

	for _, a := range actions {
		entry := getScoutEntry(state, a.ScoutFromLeft, a.ScoutFlip)
		val := entry.ActiveValue()
		if val > bestVal {
			bestVal = val
			best = a
		}
	}

	// Among same-value scouts, pick a random insert position
	var sameVal []engine.Action
	for _, a := range actions {
		entry := getScoutEntry(state, a.ScoutFromLeft, a.ScoutFlip)
		if entry.ActiveValue() == bestVal {
			sameVal = append(sameVal, a)
		}
	}
	if len(sameVal) > 0 {
		return sameVal[b.rng.Intn(len(sameVal))]
	}
	return best
}

func getScoutEntry(state *engine.GameState, fromLeft bool, flip bool) engine.HandEntry {
	var entry engine.HandEntry
	if fromLeft {
		entry = state.Table.Entries[0]
	} else {
		entry = state.Table.Entries[len(state.Table.Entries)-1]
	}
	if flip {
		entry = entry.FlippedEntry()
	}
	return entry
}
