package engine

import (
	"math/rand"
	"testing"
)

func TestGenerateDeck(t *testing.T) {
	deck := GenerateDeck()
	if len(deck) != 45 {
		t.Fatalf("deck size = %d, want 45", len(deck))
	}
	// Verify all pairs are unique and Top < Bottom
	seen := map[[2]int]bool{}
	for _, c := range deck {
		if c.Top >= c.Bottom {
			t.Errorf("card %v: Top should be < Bottom", c)
		}
		key := [2]int{c.Top, c.Bottom}
		if seen[key] {
			t.Errorf("duplicate card: %v", c)
		}
		seen[key] = true
	}
}

func TestFilterDeck3Players(t *testing.T) {
	deck := GenerateDeck()
	filtered := FilterDeck(deck, 3)
	if len(filtered) != 36 {
		t.Fatalf("3-player deck = %d cards, want 36", len(filtered))
	}
	for _, c := range filtered {
		if c.Top == 10 || c.Bottom == 10 {
			t.Errorf("3-player deck should not contain 10: %v", c)
		}
	}
}

func TestFilterDeck4Players(t *testing.T) {
	deck := GenerateDeck()
	filtered := FilterDeck(deck, 4)
	if len(filtered) != 44 {
		t.Fatalf("4-player deck = %d cards, want 44", len(filtered))
	}
}

func TestFilterDeck5Players(t *testing.T) {
	deck := GenerateDeck()
	filtered := FilterDeck(deck, 5)
	if len(filtered) != 45 {
		t.Fatalf("5-player deck = %d cards, want 45", len(filtered))
	}
}

func TestShuffleAndDeal(t *testing.T) {
	deck := FilterDeck(GenerateDeck(), 3)
	rng := rand.New(rand.NewSource(42))
	hands := ShuffleAndDeal(deck, 3, rng)

	if len(hands) != 3 {
		t.Fatalf("got %d hands, want 3", len(hands))
	}
	for i, h := range hands {
		if len(h) != 12 {
			t.Errorf("hand %d has %d cards, want 12", i, len(h))
		}
	}
}
