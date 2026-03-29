package ai

import (
	"math"
	"math/rand"
	"sort"

	"github.com/daghlny/scout_cli/pkg/engine"
)

// SmartBot uses heuristic scoring for strategic decisions:
// - "Just enough" showing (save strong combos)
// - Position-aware scouting (build combos through insert placement)
// - Defensive awareness (prevent opponents from winning rounds via consecutive scouts)
// - Strategic Scout & Show timing
type SmartBot struct {
	rng *rand.Rand
}

func NewSmartBot(rng *rand.Rand) *SmartBot {
	return &SmartBot{rng: rng}
}

func (b *SmartBot) Name() string { return "Smart" }

func (b *SmartBot) ChooseHandOrientation(hand *engine.Hand, state *engine.GameState) bool {
	original := evaluateHandQuality(hand)
	hand.FlipAll()
	flipped := evaluateHandQuality(hand)
	hand.FlipAll() // restore
	return flipped > original
}

func (b *SmartBot) ChooseAction(state *engine.GameState, playerID int) engine.Action {
	actions := engine.ValidActions(state)
	player := &state.Players[playerID]
	hand := player.Hand

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

	isTableOwner := state.Table.Combo != nil && state.Table.OwnerID == playerID
	// Defensive: if next scout ends round for opponent, must show
	mustShow := state.ConsecutiveScouts >= state.NumPlayers-2 &&
		state.Table.Combo != nil && !isTableOwner

	// Score best show
	var bestShow *engine.Action
	bestShowScore := math.Inf(-1)
	for i := range shows {
		s := b.scoreShowAction(shows[i], hand, state, playerID)
		if s > bestShowScore {
			bestShowScore = s
			bestShow = &shows[i]
		}
	}

	// Score best scout (skip if must show)
	var bestScout *engine.Action
	bestScoutScore := math.Inf(-1)
	if !mustShow && len(scouts) > 0 {
		bestScout, bestScoutScore = b.pickBestScout(scouts, hand, state, playerID)
	}

	// Evaluate Scout & Show
	bestSS, bestSSScore := b.evaluateScoutAndShow(scoutShows, hand, state, playerID)

	// === Decision ===

	if mustShow {
		if bestShow != nil {
			return *bestShow
		}
		if bestSS != nil {
			return *bestSS
		}
		// No show possible, forced to scout (rare)
		if bestScout != nil {
			return *bestScout
		}
		return actions[0]
	}

	// Can empty hand? Always do it.
	if bestShow != nil && bestShow.ShowCount == hand.Len() {
		return *bestShow
	}

	// S&S enables strong play (4+ cards or hand-emptying)
	if bestSS != nil && bestSSScore > bestShowScore && bestSSScore >= 15.0 {
		return *bestSS
	}

	// Default: prefer show if available (clears cards, collects table)
	// Only scout if it's significantly better
	if bestShow != nil {
		if bestScout != nil && bestScoutScore > bestShowScore+5.0 {
			return *bestScout
		}
		return *bestShow
	}

	if bestScout != nil {
		return *bestScout
	}
	return actions[0]
}

func (b *SmartBot) scoreShowAction(action engine.Action, hand *engine.Hand, state *engine.GameState, playerID int) float64 {
	vals, err := hand.AdjacentValues(action.ShowStart, action.ShowCount)
	if err != nil {
		return -1000
	}
	entries := hand.Entries()[action.ShowStart : action.ShowStart+action.ShowCount]
	cards := make([]engine.Card, action.ShowCount)
	for i, e := range entries {
		cards[i] = e.Card
	}
	combo, err := engine.ValidateCombo(vals, cards)
	if err != nil {
		return -1000
	}

	score := 0.0
	handSize := hand.Len()
	isTableOwner := state.Table.Combo != nil && state.Table.OwnerID == playerID

	// Base: prefer clearing more cards
	score += float64(combo.Size()) * 3.0

	// Emptying hand is the ultimate goal
	if combo.Size() == handSize {
		return 100.0
	}

	// 4+ card combos are very hard to beat → potential round-ender
	if combo.Size() >= 4 {
		score += 15.0
	}

	// "Just enough" beating: prefer minimal overkill to save strong combos
	if state.Table.Combo != nil && combo.Size() == state.Table.Combo.Size() {
		gap := combo.MinValue() - state.Table.Combo.MinValue()
		if combo.Type == engine.ComboSet && state.Table.Combo.Type == engine.ComboRun {
			gap = 1 // type upgrade counts as minimal
		}
		if gap > 0 && gap <= 2 {
			score += 5.0
		}
	}

	// Small hand: clearing bonus doubles
	if handSize <= 4 {
		score += float64(combo.Size()) * 5.0
	}

	// Defensive: if ConsecutiveScouts high and not table owner, any show is critical
	if state.ConsecutiveScouts >= state.NumPlayers-2 && !isTableOwner {
		score += 20.0
	}

	return score
}

func (b *SmartBot) pickBestScout(scouts []engine.Action, hand *engine.Hand, state *engine.GameState, playerID int) (*engine.Action, float64) {
	type scored struct {
		action engine.Action
		score  float64
	}

	var candidates []scored
	for _, a := range scouts {
		s := b.scoreScoutAction(a, hand, state, playerID)
		candidates = append(candidates, scored{a, s})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if len(candidates) == 0 {
		return nil, math.Inf(-1)
	}

	best := candidates[0]
	return &best.action, best.score
}

func (b *SmartBot) scoreScoutAction(action engine.Action, hand *engine.Hand, state *engine.GameState, playerID int) float64 {
	isTableOwner := state.Table.Combo != nil && state.Table.OwnerID == playerID

	// Defensive: never scout if it ends round for opponent
	if state.ConsecutiveScouts >= state.NumPlayers-2 && !isTableOwner {
		return -1000.0
	}

	entry := getSmartScoutEntry(state, action.ScoutFromLeft, action.ScoutFlip)
	score := scoreInsertPosition(hand, entry, action.ScoutInsertAt)

	// Slight bonus for higher-value cards
	score += float64(entry.ActiveValue()) * 0.1

	// Penalty for giving opponent a scout token
	if !isTableOwner {
		score -= 1.0
	}

	// Bonus if we are table owner (scouting = others might scout us = more tokens)
	if isTableOwner {
		score += 0.5
	}

	return score
}

func (b *SmartBot) evaluateScoutAndShow(actions []engine.Action, hand *engine.Hand, state *engine.GameState, playerID int) (*engine.Action, float64) {
	if len(actions) == 0 {
		return nil, math.Inf(-1)
	}

	var best *engine.Action
	bestScore := math.Inf(-1)

	for i := range actions {
		a := &actions[i]

		// Skip single-card shows (not worth the chip)
		if a.ShowCount < 2 {
			continue
		}

		score := float64(a.ShowCount) * 5.0

		if a.ShowCount >= 3 {
			score += 15.0
		}
		if a.ShowCount >= 4 {
			score += 25.0
		}

		// Simulate to check if it empties hand
		simHand := engine.NewHand(hand.Entries())
		entry := getSmartScoutEntry(state, a.ScoutFromLeft, a.ScoutFlip)
		_ = simHand.Insert(entry, a.ScoutInsertAt)
		if a.ShowCount == simHand.Len() {
			score += 100.0
		}

		// Penalty for using chip early
		if hand.Len() > 8 {
			score -= 10.0
		}

		if score > bestScore {
			bestScore = score
			best = a
		}
	}

	// Only recommend if above threshold
	if bestScore < 10.0 {
		return nil, math.Inf(-1)
	}
	return best, bestScore
}

// evaluateHandQuality scores a hand's combo potential for orientation choice.
func evaluateHandQuality(hand *engine.Hand) float64 {
	combos := FindAllCombos(hand)
	score := 0.0
	maxSize := 0
	covered := make(map[int]bool)

	for _, c := range combos {
		switch c.Count {
		case 1:
			score += 0.5
		case 2:
			score += 2.0
		case 3:
			score += 5.0
		default:
			score += 10.0
		}
		if c.Count >= 2 {
			for p := c.Start; p < c.Start+c.Count; p++ {
				covered[p] = true
			}
		}
		if c.Count > maxSize {
			maxSize = c.Count
		}
	}

	score += float64(len(covered)) * 1.0
	score += float64(maxSize) * 2.0
	return score
}

// scoreInsertPosition evaluates how good an insert position is using O(1) neighbor checks.
func scoreInsertPosition(hand *engine.Hand, entry engine.HandEntry, pos int) float64 {
	vals := hand.ActiveValues()
	newVal := entry.ActiveValue()
	score := 0.0

	// Check left neighbor
	if pos > 0 {
		leftVal := vals[pos-1]
		if leftVal == newVal {
			score += 3.0 // pair potential
		} else if abs(leftVal-newVal) == 1 {
			score += 2.0 // run potential
		}
	}

	// Check right neighbor (at position pos in original hand = after insert it's pos+1)
	if pos < len(vals) {
		rightVal := vals[pos]
		if rightVal == newVal {
			score += 3.0
		} else if abs(rightVal-newVal) == 1 {
			score += 2.0
		}
	}

	// Check if bridges left and right into 3-card combo
	if pos > 0 && pos < len(vals) {
		leftVal := vals[pos-1]
		rightVal := vals[pos]
		// Triple (set of 3)
		if leftVal == newVal && newVal == rightVal {
			score += 8.0
		}
		// 3-card run
		triple := []int{leftVal, newVal, rightVal}
		sort.Ints(triple)
		if triple[2]-triple[1] == 1 && triple[1]-triple[0] == 1 {
			score += 6.0
		}
	}

	return score
}

func getSmartScoutEntry(state *engine.GameState, fromLeft bool, flip bool) engine.HandEntry {
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
