package ai

import (
	"math/rand"
	"testing"

	"github.com/daghlny/scout_cli/pkg/engine"
)

func playFullGame(t *testing.T, rng *rand.Rand) {
	t.Helper()

	bots := []Strategy{
		NewGreedyBot(rand.New(rand.NewSource(rng.Int63()))),
		NewGreedyBot(rand.New(rand.NewSource(rng.Int63()))),
		NewRandomBot(rand.New(rand.NewSource(rng.Int63()))),
	}

	g, err := engine.NewGame(3, []string{"Greedy1", "Greedy2", "Random"}, rng)
	if err != nil {
		t.Fatal(err)
	}

	for round := 0; round < g.TotalRounds; round++ {
		// Setup phase: each bot chooses orientation
		for i, bot := range bots {
			flip := bot.ChooseHandOrientation(g.Players[i].Hand, g)
			if err := g.SetHandOrientation(i, flip); err != nil {
				t.Fatalf("round %d, player %d orientation: %v", round, i, err)
			}
		}
		if err := g.FinishSetup(); err != nil {
			t.Fatalf("round %d finish setup: %v", round, err)
		}

		// Play turns until round ends
		turnCount := 0
		maxTurns := 500 // safety valve
		for g.Phase == engine.PhaseTurn && turnCount < maxTurns {
			pid := g.CurrentPlayer
			action := bots[pid].ChooseAction(g, pid)
			if err := g.ApplyAction(action); err != nil {
				t.Fatalf("round %d, turn %d, player %d: %v (action: %+v)", round, turnCount, pid, err, action)
			}
			turnCount++
		}

		if turnCount >= maxTurns {
			t.Fatalf("round %d: exceeded %d turns", round, maxTurns)
		}

		if g.Phase == engine.PhaseRoundEnd {
			if err := g.NextRound(); err != nil {
				t.Fatalf("round %d next: %v", round, err)
			}
		}
	}

	if g.Phase != engine.PhaseGameEnd {
		t.Errorf("expected PhaseGameEnd, got %d", g.Phase)
	}
}

func TestBotSmokeTest(t *testing.T) {
	for i := 0; i < 100; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		playFullGame(t, rng)
	}
}

func TestBotStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	for i := 0; i < 1000; i++ {
		rng := rand.New(rand.NewSource(int64(i * 7919)))
		playFullGame(t, rng)
	}
}
