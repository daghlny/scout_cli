package engine

import "errors"

// Hand represents an ordered sequence of cards. The order cannot be rearranged;
// only insertion (via Scout) and removal (via Show) are allowed.
type Hand struct {
	entries []HandEntry
}

// NewHand creates a hand from dealt entries.
func NewHand(entries []HandEntry) *Hand {
	e := make([]HandEntry, len(entries))
	copy(e, entries)
	return &Hand{entries: e}
}

// Len returns the number of cards in hand.
func (h *Hand) Len() int {
	return len(h.entries)
}

// IsEmpty returns true if hand has no cards.
func (h *Hand) IsEmpty() bool {
	return len(h.entries) == 0
}

// Get returns the entry at the given index.
func (h *Hand) Get(index int) (HandEntry, error) {
	if index < 0 || index >= len(h.entries) {
		return HandEntry{}, errors.New("index out of range")
	}
	return h.entries[index], nil
}

// Entries returns a copy of all hand entries.
func (h *Hand) Entries() []HandEntry {
	result := make([]HandEntry, len(h.entries))
	copy(result, h.entries)
	return result
}

// ActiveValues returns all active values in hand order.
func (h *Hand) ActiveValues() []int {
	vals := make([]int, len(h.entries))
	for i, e := range h.entries {
		vals[i] = e.ActiveValue()
	}
	return vals
}

// FlipAll flips the entire hand (reverses order and toggles each card's orientation).
// This simulates physically flipping the entire fan of cards upside-down.
func (h *Hand) FlipAll() {
	n := len(h.entries)
	reversed := make([]HandEntry, n)
	for i, e := range h.entries {
		reversed[n-1-i] = e.FlippedEntry()
	}
	h.entries = reversed
}

// Insert adds a card at the given position (0 = before first, Len() = after last).
func (h *Hand) Insert(entry HandEntry, position int) error {
	if position < 0 || position > len(h.entries) {
		return errors.New("insert position out of range")
	}
	h.entries = append(h.entries, HandEntry{})
	copy(h.entries[position+1:], h.entries[position:])
	h.entries[position] = entry
	return nil
}

// RemoveRange removes `count` adjacent cards starting at `start` and returns them.
func (h *Hand) RemoveRange(start, count int) ([]HandEntry, error) {
	if start < 0 || count <= 0 || start+count > len(h.entries) {
		return nil, errors.New("remove range out of bounds")
	}
	removed := make([]HandEntry, count)
	copy(removed, h.entries[start:start+count])
	h.entries = append(h.entries[:start], h.entries[start+count:]...)
	return removed, nil
}

// AdjacentValues returns active values for `count` cards starting at `start`.
func (h *Hand) AdjacentValues(start, count int) ([]int, error) {
	if start < 0 || count <= 0 || start+count > len(h.entries) {
		return nil, errors.New("range out of bounds")
	}
	vals := make([]int, count)
	for i := 0; i < count; i++ {
		vals[i] = h.entries[start+i].ActiveValue()
	}
	return vals, nil
}
