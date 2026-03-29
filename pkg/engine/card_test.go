package engine

import "testing"

func TestNewCard(t *testing.T) {
	c := NewCard(7, 3)
	if c.Top != 3 || c.Bottom != 7 {
		t.Errorf("NewCard(7,3) = %v, want Top=3, Bottom=7", c)
	}

	c2 := NewCard(2, 5)
	if c2.Top != 2 || c2.Bottom != 5 {
		t.Errorf("NewCard(2,5) = %v, want Top=2, Bottom=5", c2)
	}
}

func TestHandEntryActiveValue(t *testing.T) {
	c := Card{Top: 3, Bottom: 7}
	e := HandEntry{Card: c, Flipped: false}
	if e.ActiveValue() != 3 {
		t.Errorf("ActiveValue (not flipped) = %d, want 3", e.ActiveValue())
	}

	e.Flipped = true
	if e.ActiveValue() != 7 {
		t.Errorf("ActiveValue (flipped) = %d, want 7", e.ActiveValue())
	}
}

func TestHandEntryFlipped(t *testing.T) {
	e := HandEntry{Card: Card{Top: 1, Bottom: 9}, Flipped: false}
	f := e.FlippedEntry()
	if f.Flipped != true || f.ActiveValue() != 9 {
		t.Errorf("FlippedEntry: flipped=%v, active=%d", f.Flipped, f.ActiveValue())
	}
}
