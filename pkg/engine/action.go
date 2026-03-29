package engine

// ActionType represents the type of action a player can take on their turn.
type ActionType int

const (
	ActionShow         ActionType = iota // Play cards from hand to beat the table
	ActionScout                          // Take a card from either end of the table combo
	ActionScoutAndShow                   // Scout then Show (once per round)
)

func (t ActionType) String() string {
	switch t {
	case ActionShow:
		return "Show"
	case ActionScout:
		return "Scout"
	case ActionScoutAndShow:
		return "Scout & Show"
	default:
		return "Unknown"
	}
}

// Action represents a fully specified player action.
type Action struct {
	Type ActionType

	// Show fields (also used for ScoutAndShow's show phase)
	ShowStart int // start index in hand
	ShowCount int // number of adjacent cards to play

	// Scout fields (also used for ScoutAndShow's scout phase)
	ScoutFromLeft bool // true: take leftmost card, false: take rightmost
	ScoutFlip     bool // flip the scouted card before inserting
	ScoutInsertAt int  // position in hand to insert the scouted card
}

// ValidActions generates all legal actions for the current player given game state.
func ValidActions(state *GameState) []Action {
	player := &state.Players[state.CurrentPlayer]
	hand := player.Hand
	var actions []Action

	tableEmpty := state.Table.Combo == nil

	// === Show actions ===
	// Try every adjacent slice in hand
	for start := 0; start < hand.Len(); start++ {
		for count := 1; count <= hand.Len()-start; count++ {
			values, _ := hand.AdjacentValues(start, count)
			entries := hand.Entries()[start : start+count]
			cards := make([]Card, count)
			for i, e := range entries {
				cards[i] = e.Card
			}
			combo, err := ValidateCombo(values, cards)
			if err != nil {
				continue
			}
			if !tableEmpty && !combo.Beats(state.Table.Combo) {
				continue
			}
			actions = append(actions, Action{
				Type:      ActionShow,
				ShowStart: start,
				ShowCount: count,
			})
		}
	}

	// === Scout actions (only if table is not empty) ===
	if !tableEmpty && state.Table.Combo.Size() > 0 {
		for _, fromLeft := range []bool{true, false} {
			// If combo has only 1 card, left and right are the same; skip duplicate
			if state.Table.Combo.Size() == 1 && !fromLeft {
				continue
			}
			for _, flip := range []bool{false, true} {
				for pos := 0; pos <= hand.Len(); pos++ {
					actions = append(actions, Action{
						Type:          ActionScout,
						ScoutFromLeft: fromLeft,
						ScoutFlip:     flip,
						ScoutInsertAt: pos,
					})
				}
			}
		}
	}

	// === Scout & Show actions (only if chip not used and table not empty) ===
	if !player.UsedScoutShow && !tableEmpty && state.Table.Combo.Size() > 0 {
		// For each scout option, simulate the scout, then find valid shows
		for _, fromLeft := range []bool{true, false} {
			if state.Table.Combo.Size() == 1 && !fromLeft {
				continue
			}
			for _, flip := range []bool{false, true} {
				for insertPos := 0; insertPos <= hand.Len(); insertPos++ {
					// Simulate the scout
					simHand := NewHand(hand.Entries())
					scoutedEntry := scoutEntry(state.Table, fromLeft, flip)
					_ = simHand.Insert(scoutedEntry, insertPos)

					// Determine the reduced table combo after scouting
					reducedCombo := reducedTableCombo(state.Table, fromLeft)

					// Find all valid shows from the simulated hand
					for showStart := 0; showStart < simHand.Len(); showStart++ {
						for showCount := 1; showCount <= simHand.Len()-showStart; showCount++ {
							vals, _ := simHand.AdjacentValues(showStart, showCount)
							entries := simHand.Entries()[showStart : showStart+showCount]
							cards := make([]Card, showCount)
							for i, e := range entries {
								cards[i] = e.Card
							}
							combo, err := ValidateCombo(vals, cards)
							if err != nil {
								continue
							}
							// Must beat the reduced table combo (or table is now empty)
							if reducedCombo != nil && !combo.Beats(reducedCombo) {
								continue
							}
							actions = append(actions, Action{
								Type:          ActionScoutAndShow,
								ScoutFromLeft: fromLeft,
								ScoutFlip:     flip,
								ScoutInsertAt: insertPos,
								ShowStart:     showStart,
								ShowCount:     showCount,
							})
						}
					}
				}
			}
		}
	}

	return actions
}

// scoutEntry creates the HandEntry that would result from scouting.
func scoutEntry(table TableState, fromLeft bool, flip bool) HandEntry {
	var idx int
	if fromLeft {
		idx = 0
	} else {
		idx = len(table.Entries) - 1
	}
	entry := table.Entries[idx]
	if flip {
		entry = entry.FlippedEntry()
	}
	return entry
}

// reducedTableCombo returns the combo remaining after scouting one card from an end.
// Returns nil if the combo is reduced to zero cards.
func reducedTableCombo(table TableState, fromLeft bool) *Combo {
	if table.Combo.Size() <= 1 {
		return nil
	}

	var remainEntries []HandEntry
	if fromLeft {
		remainEntries = table.Entries[1:]
	} else {
		remainEntries = table.Entries[:len(table.Entries)-1]
	}

	values := make([]int, len(remainEntries))
	cards := make([]Card, len(remainEntries))
	for i, e := range remainEntries {
		values[i] = e.ActiveValue()
		cards[i] = e.Card
	}

	combo, err := ValidateCombo(values, cards)
	if err != nil {
		// After removing an end card, the remaining may not form a valid combo,
		// but it still stays on the table as-is. Treat it as the existing combo minus one.
		// For comparison purposes, we need the original combo's reduced form.
		return nil
	}
	return combo
}
