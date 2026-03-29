package engine

import (
	"errors"
	"fmt"
	"math/rand"
)

// Phase represents the current phase of the game.
type Phase int

const (
	PhaseSetup      Phase = iota // waiting for hand orientation choices
	PhaseTurn                    // active player chooses an action
	PhaseRoundEnd                // round just ended, scores calculated
	PhaseGameEnd                 // all rounds complete
)

// TableState represents the current cards on the table.
type TableState struct {
	Combo   *Combo      // nil if table is empty
	OwnerID int         // who played the current combo
	Entries []HandEntry // physical entries on table (for scouting from ends)
}

// GameState is the complete game state, driven by ApplyAction.
type GameState struct {
	Phase             Phase
	NumPlayers        int
	Players           []PlayerState
	CurrentPlayer     int
	Table             TableState
	Round             int // 0-indexed current round
	TotalRounds       int
	StartPlayer       int // who starts the current round
	ConsecutiveScouts int // consecutive non-Show actions since last Show
	RoundEnderID      int // who ended the round (-1 if round not ended)

	rng *rand.Rand
}

// NewGame creates a new game for the given number of players (3-5).
func NewGame(numPlayers int, playerNames []string, rng *rand.Rand) (*GameState, error) {
	if numPlayers < 3 || numPlayers > 5 {
		return nil, fmt.Errorf("invalid player count: %d (must be 3-5)", numPlayers)
	}
	if len(playerNames) != numPlayers {
		return nil, errors.New("player names count must match player count")
	}

	g := &GameState{
		Phase:        PhaseSetup,
		NumPlayers:   numPlayers,
		TotalRounds:  numPlayers,
		Round:        0,
		StartPlayer:  0,
		RoundEnderID: -1,
		rng:          rng,
	}

	g.Players = make([]PlayerState, numPlayers)
	for i := 0; i < numPlayers; i++ {
		g.Players[i] = PlayerState{
			ID:   i,
			Name: playerNames[i],
		}
	}

	g.dealRound()
	return g, nil
}

// dealRound shuffles, filters, and deals cards for the current round.
func (g *GameState) dealRound() {
	deck := GenerateDeck()
	filtered := FilterDeck(deck, g.NumPlayers)
	hands := ShuffleAndDeal(filtered, g.NumPlayers, g.rng)

	for i := range g.Players {
		g.Players[i].ResetForRound()
		g.Players[i].Hand = NewHand(hands[i])
	}

	g.Table = TableState{}
	g.CurrentPlayer = g.StartPlayer
	g.ConsecutiveScouts = 0
	g.RoundEnderID = -1
	g.Phase = PhaseSetup
}

// SetHandOrientation records a player's flip-or-keep decision.
// Call for each player during PhaseSetup. Once all have decided, advances to PhaseTurn.
func (g *GameState) SetHandOrientation(playerID int, flip bool) error {
	if g.Phase != PhaseSetup {
		return errors.New("not in setup phase")
	}
	if playerID < 0 || playerID >= g.NumPlayers {
		return errors.New("invalid player ID")
	}
	if flip {
		g.Players[playerID].Hand.FlipAll()
	}
	return nil
}

// FinishSetup transitions from PhaseSetup to PhaseTurn after all orientations are set.
func (g *GameState) FinishSetup() error {
	if g.Phase != PhaseSetup {
		return errors.New("not in setup phase")
	}
	g.Phase = PhaseTurn
	return nil
}

// ApplyAction is the core state machine transition. It validates and executes
// the action for the current player, then advances the game state.
func (g *GameState) ApplyAction(action Action) error {
	if g.Phase != PhaseTurn {
		return fmt.Errorf("cannot apply action in phase %d", g.Phase)
	}

	player := &g.Players[g.CurrentPlayer]
	tableEmpty := g.Table.Combo == nil

	// First player of a round must Show (nothing to Scout from)
	if tableEmpty && action.Type != ActionShow {
		return errors.New("must Show when table is empty")
	}

	switch action.Type {
	case ActionShow:
		return g.applyShow(player, action)
	case ActionScout:
		if tableEmpty {
			return errors.New("cannot Scout when table is empty")
		}
		return g.applyScout(player, action)
	case ActionScoutAndShow:
		if tableEmpty {
			return errors.New("cannot Scout & Show when table is empty")
		}
		if player.UsedScoutShow {
			return errors.New("Scout & Show already used this round")
		}
		return g.applyScoutAndShow(player, action)
	default:
		return fmt.Errorf("unknown action type: %d", action.Type)
	}
}

func (g *GameState) applyShow(player *PlayerState, action Action) error {
	// Extract cards from hand
	values, err := player.Hand.AdjacentValues(action.ShowStart, action.ShowCount)
	if err != nil {
		return fmt.Errorf("invalid show range: %w", err)
	}

	entries := player.Hand.Entries()[action.ShowStart : action.ShowStart+action.ShowCount]
	cards := make([]Card, action.ShowCount)
	for i, e := range entries {
		cards[i] = e.Card
	}

	combo, err := ValidateCombo(values, cards)
	if err != nil {
		return fmt.Errorf("invalid combo: %w", err)
	}

	// Must beat current table combo (if any)
	if g.Table.Combo != nil && !combo.Beats(g.Table.Combo) {
		return errors.New("combo does not beat the table")
	}

	// Collect previous table cards
	if g.Table.Combo != nil {
		for _, e := range g.Table.Entries {
			player.CollectedCards = append(player.CollectedCards, e.Card)
		}
	}

	// Remove cards from hand
	showEntries, _ := player.Hand.RemoveRange(action.ShowStart, action.ShowCount)

	// Place new combo on table
	g.Table = TableState{
		Combo:   combo,
		OwnerID: player.ID,
		Entries: showEntries,
	}

	g.ConsecutiveScouts = 0

	// Check if player emptied their hand
	if player.Hand.IsEmpty() {
		return g.endRound(player.ID)
	}

	g.advanceTurn()
	return nil
}

func (g *GameState) applyScout(player *PlayerState, action Action) error {
	if g.Table.Combo == nil || len(g.Table.Entries) == 0 {
		return errors.New("nothing to scout from")
	}

	// Take card from table end
	entry := g.takeFromTable(action.ScoutFromLeft)
	if action.ScoutFlip {
		entry = entry.FlippedEntry()
	}

	// Insert into hand
	if err := player.Hand.Insert(entry, action.ScoutInsertAt); err != nil {
		return fmt.Errorf("invalid insert position: %w", err)
	}

	// Owner gets a scout token
	g.Players[g.Table.OwnerID].ScoutTokens++

	g.ConsecutiveScouts++

	// Check if all other players have scouted (round end condition)
	if g.ConsecutiveScouts >= g.NumPlayers-1 {
		return g.endRound(g.Table.OwnerID)
	}

	g.advanceTurn()
	return nil
}

func (g *GameState) applyScoutAndShow(player *PlayerState, action Action) error {
	// Mark chip as used
	player.UsedScoutShow = true

	// === Scout phase ===
	entry := g.takeFromTable(action.ScoutFromLeft)
	if action.ScoutFlip {
		entry = entry.FlippedEntry()
	}
	if err := player.Hand.Insert(entry, action.ScoutInsertAt); err != nil {
		return fmt.Errorf("invalid insert position: %w", err)
	}
	g.Players[g.Table.OwnerID].ScoutTokens++

	// === Show phase ===
	values, err := player.Hand.AdjacentValues(action.ShowStart, action.ShowCount)
	if err != nil {
		return fmt.Errorf("invalid show range: %w", err)
	}

	showEntries := player.Hand.Entries()[action.ShowStart : action.ShowStart+action.ShowCount]
	cards := make([]Card, action.ShowCount)
	for i, e := range showEntries {
		cards[i] = e.Card
	}

	combo, err := ValidateCombo(values, cards)
	if err != nil {
		return fmt.Errorf("invalid combo: %w", err)
	}

	// Must beat reduced table combo (if table still has cards)
	if g.Table.Combo != nil && !combo.Beats(g.Table.Combo) {
		return errors.New("combo does not beat the table after scout")
	}

	// Collect remaining table cards
	if g.Table.Combo != nil {
		for _, e := range g.Table.Entries {
			player.CollectedCards = append(player.CollectedCards, e.Card)
		}
	}

	// Remove shown cards from hand
	removedEntries, _ := player.Hand.RemoveRange(action.ShowStart, action.ShowCount)

	// Place new combo on table
	g.Table = TableState{
		Combo:   combo,
		OwnerID: player.ID,
		Entries: removedEntries,
	}

	g.ConsecutiveScouts = 0

	if player.Hand.IsEmpty() {
		return g.endRound(player.ID)
	}

	g.advanceTurn()
	return nil
}

// takeFromTable removes a card from the left or right end of the table combo.
func (g *GameState) takeFromTable(fromLeft bool) HandEntry {
	var entry HandEntry
	if fromLeft {
		entry = g.Table.Entries[0]
		g.Table.Entries = g.Table.Entries[1:]
	} else {
		last := len(g.Table.Entries) - 1
		entry = g.Table.Entries[last]
		g.Table.Entries = g.Table.Entries[:last]
	}

	// Update table combo
	if len(g.Table.Entries) == 0 {
		g.Table.Combo = nil
	} else {
		values := make([]int, len(g.Table.Entries))
		cards := make([]Card, len(g.Table.Entries))
		for i, e := range g.Table.Entries {
			values[i] = e.ActiveValue()
			cards[i] = e.Card
		}
		newCombo, err := ValidateCombo(values, cards)
		if err != nil {
			// Remaining cards may not form a valid combo after removal.
			// Keep the entries on table but mark combo as nil for comparison purposes.
			g.Table.Combo = nil
		} else {
			g.Table.Combo = newCombo
		}
	}

	return entry
}

// advanceTurn moves to the next player.
func (g *GameState) advanceTurn() {
	g.CurrentPlayer = (g.CurrentPlayer + 1) % g.NumPlayers
}

// endRound handles round-end scoring and transitions.
func (g *GameState) endRound(enderID int) error {
	g.RoundEnderID = enderID
	g.Phase = PhaseRoundEnd

	scores := ScoreRound(g.Players, enderID)
	for i := range g.Players {
		g.Players[i].RoundScores = append(g.Players[i].RoundScores, scores[i])
	}

	return nil
}

// NextRound advances to the next round, or ends the game.
func (g *GameState) NextRound() error {
	if g.Phase != PhaseRoundEnd {
		return errors.New("not in round end phase")
	}

	g.Round++
	if g.Round >= g.TotalRounds {
		g.Phase = PhaseGameEnd
		return nil
	}

	g.StartPlayer = (g.StartPlayer + 1) % g.NumPlayers
	g.dealRound()
	return nil
}

// Winner returns the player ID with the highest total score. Ties go to lower ID.
func (g *GameState) Winner() int {
	bestID := 0
	bestScore := g.Players[0].TotalScore()
	for i := 1; i < len(g.Players); i++ {
		s := g.Players[i].TotalScore()
		if s > bestScore {
			bestScore = s
			bestID = i
		}
	}
	return bestID
}
