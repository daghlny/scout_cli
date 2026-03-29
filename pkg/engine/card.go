package engine

import "fmt"

// Card represents a double-sided Scout card with two different numbers.
type Card struct {
	Top    int
	Bottom int
}

// HandEntry pairs a card with its current orientation in a player's hand.
type HandEntry struct {
	Card    Card
	Flipped bool // false: Top is active, true: Bottom is active
}

// ActiveValue returns the number currently facing up based on orientation.
func (e HandEntry) ActiveValue() int {
	if e.Flipped {
		return e.Card.Bottom
	}
	return e.Card.Top
}

// InactiveValue returns the number on the hidden side.
func (e HandEntry) InactiveValue() int {
	if e.Flipped {
		return e.Card.Top
	}
	return e.Card.Bottom
}

// String returns a display string like "3|7" (active|inactive).
func (e HandEntry) String() string {
	return fmt.Sprintf("%d|%d", e.ActiveValue(), e.InactiveValue())
}

// FlippedEntry returns a new HandEntry with the opposite orientation.
func (e HandEntry) FlippedEntry() HandEntry {
	return HandEntry{Card: e.Card, Flipped: !e.Flipped}
}

// NewCard creates a card ensuring Top < Bottom for canonical representation.
func NewCard(a, b int) Card {
	if a < b {
		return Card{Top: a, Bottom: b}
	}
	return Card{Top: b, Bottom: a}
}

func (c Card) String() string {
	return fmt.Sprintf("[%d/%d]", c.Top, c.Bottom)
}
