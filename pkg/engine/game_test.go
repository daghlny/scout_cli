package engine

import (
	"math/rand"
	"testing"
)

func newTestGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(42))
	g, err := NewGame(3, []string{"Alice", "Bot1", "Bot2"}, rng)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestNewGameSetup(t *testing.T) {
	g := newTestGame(t)
	if g.Phase != PhaseSetup {
		t.Errorf("phase = %d, want PhaseSetup", g.Phase)
	}
	if g.NumPlayers != 3 {
		t.Errorf("numPlayers = %d, want 3", g.NumPlayers)
	}
	for i, p := range g.Players {
		if p.Hand.Len() != 12 {
			t.Errorf("player %d hand = %d cards, want 12", i, p.Hand.Len())
		}
	}
}

func TestSetupAndFirstShow(t *testing.T) {
	g := newTestGame(t)

	// All players keep their hand orientation
	for i := 0; i < 3; i++ {
		if err := g.SetHandOrientation(i, false); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.FinishSetup(); err != nil {
		t.Fatal(err)
	}

	if g.Phase != PhaseTurn {
		t.Fatalf("phase = %d, want PhaseTurn", g.Phase)
	}

	// First player must Show (table is empty)
	// Find a valid single-card show
	action := Action{
		Type:      ActionShow,
		ShowStart: 0,
		ShowCount: 1,
	}
	if err := g.ApplyAction(action); err != nil {
		t.Fatal(err)
	}

	if g.Table.Combo == nil {
		t.Fatal("table should have a combo after Show")
	}
	if g.CurrentPlayer != 1 {
		t.Errorf("current player = %d, want 1", g.CurrentPlayer)
	}
}

func TestScoutAction(t *testing.T) {
	g := newTestGame(t)
	for i := 0; i < 3; i++ {
		_ = g.SetHandOrientation(i, false)
	}
	_ = g.FinishSetup()

	// Player 0 shows a single card
	_ = g.ApplyAction(Action{Type: ActionShow, ShowStart: 0, ShowCount: 1})

	// Player 1 scouts from left end
	handSizeBefore := g.Players[1].Hand.Len()
	action := Action{
		Type:          ActionScout,
		ScoutFromLeft: true,
		ScoutFlip:     false,
		ScoutInsertAt: 0,
	}
	if err := g.ApplyAction(action); err != nil {
		t.Fatal(err)
	}

	if g.Players[1].Hand.Len() != handSizeBefore+1 {
		t.Errorf("hand size after scout = %d, want %d", g.Players[1].Hand.Len(), handSizeBefore+1)
	}
	if g.Players[0].ScoutTokens != 1 {
		t.Errorf("scout tokens = %d, want 1", g.Players[0].ScoutTokens)
	}
}

func TestRoundEndByConsecutiveScouts(t *testing.T) {
	g := newTestGame(t)
	for i := 0; i < 3; i++ {
		_ = g.SetHandOrientation(i, false)
	}
	_ = g.FinishSetup()

	// Player 0 shows a 2-card combo so scouts don't empty the table
	// First find two adjacent cards that form a valid combo
	hand := g.Players[0].Hand
	var showStart, showCount int
	found := false
	for s := 0; s < hand.Len()-1; s++ {
		vals, _ := hand.AdjacentValues(s, 2)
		entries := hand.Entries()[s : s+2]
		cs := []Card{entries[0].Card, entries[1].Card}
		_, err := ValidateCombo(vals, cs)
		if err == nil {
			showStart = s
			showCount = 2
			found = true
			break
		}
	}
	if !found {
		// Fallback: just show a single card, player 1 shows, then 2 scouts
		_ = g.ApplyAction(Action{Type: ActionShow, ShowStart: 0, ShowCount: 1})
		// Player 1 shows a single card (must beat player 0's card)
		actions := ValidActions(g)
		var showAction *Action
		for _, a := range actions {
			if a.Type == ActionShow {
				showAction = &a
				break
			}
		}
		if showAction == nil {
			t.Skip("no valid show action for player 1, skipping test")
		}
		_ = g.ApplyAction(*showAction)

		// Player 2 scouts
		_ = g.ApplyAction(Action{Type: ActionScout, ScoutFromLeft: true, ScoutInsertAt: 0})
		// Player 0 scouts -> 2 consecutive scouts, round ends
		err := g.ApplyAction(Action{Type: ActionScout, ScoutFromLeft: true, ScoutInsertAt: 0})
		if err != nil {
			t.Fatal(err)
		}
		if g.Phase != PhaseRoundEnd {
			t.Errorf("phase = %d, want PhaseRoundEnd", g.Phase)
		}
		return
	}

	_ = g.ApplyAction(Action{Type: ActionShow, ShowStart: showStart, ShowCount: showCount})

	// Player 1 scouts from left
	_ = g.ApplyAction(Action{Type: ActionScout, ScoutFromLeft: true, ScoutInsertAt: 0})

	// Player 2 scouts -> 2 consecutive scouts = NumPlayers-1, round should end
	err := g.ApplyAction(Action{Type: ActionScout, ScoutFromLeft: true, ScoutInsertAt: 0})
	if err != nil {
		t.Fatal(err)
	}

	if g.Phase != PhaseRoundEnd {
		t.Errorf("phase = %d, want PhaseRoundEnd", g.Phase)
	}
	if g.RoundEnderID != 0 {
		t.Errorf("round ender = %d, want 0 (the show player)", g.RoundEnderID)
	}
}

func TestScoring(t *testing.T) {
	players := []PlayerState{
		{ID: 0, Hand: NewHand(nil), CollectedCards: []Card{{1, 2}, {3, 4}}, ScoutTokens: 1},
		{ID: 1, Hand: makeTestHand([][2]int{{1, 2}, {3, 4}, {5, 6}}), CollectedCards: []Card{{7, 8}}, ScoutTokens: 0},
	}

	scores := ScoreRound(players, 0)
	// Player 0: 2 collected + 1 token - 0 hand (exempt) = 3
	if scores[0] != 3 {
		t.Errorf("player 0 score = %d, want 3", scores[0])
	}
	// Player 1: 1 collected + 0 token - 3 hand = -2
	if scores[1] != -2 {
		t.Errorf("player 1 score = %d, want -2", scores[1])
	}
}
