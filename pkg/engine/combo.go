package engine

import (
	"errors"
	"fmt"
)

// ComboType represents the type of a played card combination.
type ComboType int

const (
	ComboSingle ComboType = iota
	ComboSet              // all same number
	ComboRun              // consecutive ascending or descending
)

func (t ComboType) String() string {
	switch t {
	case ComboSingle:
		return "Single"
	case ComboSet:
		return "Set"
	case ComboRun:
		return "Run"
	default:
		return "Unknown"
	}
}

// Combo represents a valid played combination of cards.
type Combo struct {
	Type   ComboType
	Values []int  // active values in played order
	Cards  []Card // original cards (for scoring/display)
}

// Size returns the number of cards in the combo.
func (c *Combo) Size() int {
	return len(c.Values)
}

// MinValue returns the minimum value in the combo.
func (c *Combo) MinValue() int {
	min := c.Values[0]
	for _, v := range c.Values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// Beats returns true if this combo strictly beats the other combo.
// Rules: 1) more cards > fewer cards
//        2) same count: Set > Run
//        3) same count & type: higher MinValue wins
func (c *Combo) Beats(other *Combo) bool {
	if c.Size() != other.Size() {
		return c.Size() > other.Size()
	}
	if c.Type != other.Type {
		// Set beats Run (Single never has same-count conflicts with Set/Run since it's always size 1)
		return c.Type == ComboSet && other.Type == ComboRun
	}
	return c.MinValue() > other.MinValue()
}

// ValidateCombo checks if a slice of values forms a legal combination.
// Returns the Combo if valid, or an error.
func ValidateCombo(values []int, cards []Card) (*Combo, error) {
	if len(values) == 0 {
		return nil, errors.New("empty combo")
	}

	if len(values) != len(cards) {
		return nil, errors.New("values and cards length mismatch")
	}

	if len(values) == 1 {
		return &Combo{
			Type:   ComboSingle,
			Values: values,
			Cards:  cards,
		}, nil
	}

	// Check if it's a Set (all same value)
	if isSet(values) {
		return &Combo{
			Type:   ComboSet,
			Values: values,
			Cards:  cards,
		}, nil
	}

	// Check if it's a Run (consecutive ascending or descending)
	if isRun(values) {
		return &Combo{
			Type:   ComboRun,
			Values: values,
			Cards:  cards,
		}, nil
	}

	return nil, fmt.Errorf("invalid combo: values %v are neither a set nor a run", values)
}

// isSet checks if all values are the same.
func isSet(values []int) bool {
	for i := 1; i < len(values); i++ {
		if values[i] != values[0] {
			return false
		}
	}
	return true
}

// isRun checks if values form a strictly consecutive sequence (ascending or descending).
func isRun(values []int) bool {
	if len(values) < 2 {
		return false
	}

	// Check ascending: each value is exactly prev+1
	ascending := true
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			ascending = false
			break
		}
	}
	if ascending {
		return true
	}

	// Check descending: each value is exactly prev-1
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]-1 {
			return false
		}
	}
	return true
}
