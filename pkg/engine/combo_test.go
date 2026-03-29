package engine

import "testing"

func cards(n int) []Card {
	c := make([]Card, n)
	for i := range c {
		c[i] = Card{Top: 1, Bottom: 2}
	}
	return c
}

func TestValidateComboSingle(t *testing.T) {
	c, err := ValidateCombo([]int{5}, cards(1))
	if err != nil {
		t.Fatal(err)
	}
	if c.Type != ComboSingle {
		t.Errorf("type = %v, want Single", c.Type)
	}
}

func TestValidateComboSet(t *testing.T) {
	c, err := ValidateCombo([]int{7, 7, 7}, cards(3))
	if err != nil {
		t.Fatal(err)
	}
	if c.Type != ComboSet {
		t.Errorf("type = %v, want Set", c.Type)
	}
}

func TestValidateComboRunAscending(t *testing.T) {
	c, err := ValidateCombo([]int{3, 4, 5}, cards(3))
	if err != nil {
		t.Fatal(err)
	}
	if c.Type != ComboRun {
		t.Errorf("type = %v, want Run", c.Type)
	}
}

func TestValidateComboRunDescending(t *testing.T) {
	c, err := ValidateCombo([]int{8, 7, 6}, cards(3))
	if err != nil {
		t.Fatal(err)
	}
	if c.Type != ComboRun {
		t.Errorf("type = %v, want Run", c.Type)
	}
}

func TestValidateComboInvalid(t *testing.T) {
	_, err := ValidateCombo([]int{3, 5, 7}, cards(3))
	if err == nil {
		t.Error("expected error for invalid combo")
	}
}

func TestBeatsMoreCards(t *testing.T) {
	a, _ := ValidateCombo([]int{2, 3, 4}, cards(3))
	b, _ := ValidateCombo([]int{9, 9}, cards(2))
	if !a.Beats(b) {
		t.Error("3 cards should beat 2 cards")
	}
}

func TestBeatsSetOverRun(t *testing.T) {
	set, _ := ValidateCombo([]int{2, 2}, cards(2))
	run, _ := ValidateCombo([]int{9, 8}, cards(2))
	if !set.Beats(run) {
		t.Error("set should beat run of same size")
	}
}

func TestBeatsSameTypeHigherMin(t *testing.T) {
	a, _ := ValidateCombo([]int{5, 6, 7}, cards(3))
	b, _ := ValidateCombo([]int{3, 4, 5}, cards(3))
	if !a.Beats(b) {
		t.Error("run with min 5 should beat run with min 3")
	}

	c, _ := ValidateCombo([]int{8, 8}, cards(2))
	d, _ := ValidateCombo([]int{3, 3}, cards(2))
	if !c.Beats(d) {
		t.Error("set of 8s should beat set of 3s")
	}
}

func TestBeatsNotEqual(t *testing.T) {
	a, _ := ValidateCombo([]int{5, 5}, cards(2))
	b, _ := ValidateCombo([]int{5, 5}, cards(2))
	if a.Beats(b) {
		t.Error("identical combos should not beat each other")
	}
}
