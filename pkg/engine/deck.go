package engine

import "math/rand"

// GenerateDeck creates the full 45-card Scout deck (all C(10,2) pairs).
func GenerateDeck() []Card {
	var deck []Card
	for i := 1; i <= 10; i++ {
		for j := i + 1; j <= 10; j++ {
			deck = append(deck, Card{Top: i, Bottom: j})
		}
	}
	return deck
}

// FilterDeck removes cards based on player count rules.
//   - 3 players: remove all cards containing 10 (9 cards removed, 36 remain)
//   - 4 players: remove only the 9/10 card (1 removed, 44 remain)
//   - 5 players: use all 45 cards
func FilterDeck(deck []Card, numPlayers int) []Card {
	switch numPlayers {
	case 3:
		var filtered []Card
		for _, c := range deck {
			if c.Top != 10 && c.Bottom != 10 {
				filtered = append(filtered, c)
			}
		}
		return filtered
	case 4:
		var filtered []Card
		for _, c := range deck {
			if !(c.Top == 9 && c.Bottom == 10) && !(c.Top == 10 && c.Bottom == 9) {
				filtered = append(filtered, c)
			}
		}
		return filtered
	default:
		result := make([]Card, len(deck))
		copy(result, deck)
		return result
	}
}

// ShuffleAndDeal shuffles the deck with random orientations and deals to players.
// Returns a slice of hands (one per player). Each hand preserves dealt order.
func ShuffleAndDeal(deck []Card, numPlayers int, rng *rand.Rand) [][]HandEntry {
	// Shuffle card order
	shuffled := make([]Card, len(deck))
	copy(shuffled, deck)
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Create hand entries with random orientations
	entries := make([]HandEntry, len(shuffled))
	for i, c := range shuffled {
		entries[i] = HandEntry{
			Card:    c,
			Flipped: rng.Intn(2) == 1,
		}
	}

	// Deal evenly
	cardsPerPlayer := len(entries) / numPlayers
	hands := make([][]HandEntry, numPlayers)
	for i := 0; i < numPlayers; i++ {
		start := i * cardsPerPlayer
		hands[i] = make([]HandEntry, cardsPerPlayer)
		copy(hands[i], entries[start:start+cardsPerPlayer])
	}

	return hands
}

// CardsPerPlayer returns how many cards each player gets for a given player count.
func CardsPerPlayer(numPlayers int) int {
	switch numPlayers {
	case 3:
		return 12
	case 4:
		return 11
	case 5:
		return 9
	default:
		return 0
	}
}
