package engine

import (
	"testing"
)

func makeTestHand(values [][2]int) *Hand {
	entries := make([]HandEntry, len(values))
	for i, v := range values {
		entries[i] = HandEntry{Card: Card{Top: v[0], Bottom: v[1]}, Flipped: false}
	}
	return NewHand(entries)
}

func TestHandActiveValues(t *testing.T) {
	h := makeTestHand([][2]int{{3, 7}, {5, 2}, {8, 1}})
	vals := h.ActiveValues()
	expected := []int{3, 5, 8}
	for i, v := range vals {
		if v != expected[i] {
			t.Errorf("ActiveValues[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestHandFlipAll(t *testing.T) {
	h := makeTestHand([][2]int{{3, 7}, {5, 2}, {8, 1}})
	h.FlipAll()

	// After flip: reversed order, each card flipped
	// Original: [3|7, 5|2, 8|1] -> Reversed+flipped: [1|8, 2|5, 7|3]
	expected := []int{1, 2, 7}
	vals := h.ActiveValues()
	for i, v := range vals {
		if v != expected[i] {
			t.Errorf("after FlipAll: ActiveValues[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestHandInsert(t *testing.T) {
	h := makeTestHand([][2]int{{3, 7}, {5, 2}})
	newEntry := HandEntry{Card: Card{Top: 9, Bottom: 4}, Flipped: false}

	if err := h.Insert(newEntry, 1); err != nil {
		t.Fatal(err)
	}

	if h.Len() != 3 {
		t.Fatalf("Len = %d, want 3", h.Len())
	}

	vals := h.ActiveValues()
	expected := []int{3, 9, 5}
	for i, v := range vals {
		if v != expected[i] {
			t.Errorf("ActiveValues[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestHandRemoveRange(t *testing.T) {
	h := makeTestHand([][2]int{{3, 7}, {5, 2}, {8, 1}, {4, 6}})

	removed, err := h.RemoveRange(1, 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(removed) != 2 {
		t.Fatalf("removed %d, want 2", len(removed))
	}
	if removed[0].ActiveValue() != 5 || removed[1].ActiveValue() != 8 {
		t.Errorf("removed values: %d, %d", removed[0].ActiveValue(), removed[1].ActiveValue())
	}

	if h.Len() != 2 {
		t.Fatalf("remaining hand Len = %d, want 2", h.Len())
	}

	vals := h.ActiveValues()
	if vals[0] != 3 || vals[1] != 4 {
		t.Errorf("remaining values: %v, want [3, 4]", vals)
	}
}
