package sudoku

import "testing"

func TestSolve(t *testing.T) {
	board := Board{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}

	solved, ok := board.Solved()
	if !ok {
		t.Fatal("expected board to be solvable")
	}

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if solved.Get(r, c) == 0 {
				t.Fatalf("cell (%d,%d) is empty in solution", r, c)
			}
			if board.Get(r, c) != 0 && board.Get(r, c) != solved.Get(r, c) {
				t.Fatalf("given clue changed at (%d,%d)", r, c)
			}
		}
	}

	for r := 0; r < 9; r++ {
		seen := [10]bool{}
		for c := 0; c < 9; c++ {
			v := solved.Get(r, c)
			if seen[v] {
				t.Fatalf("duplicate %d in row %d", v, r)
			}
			seen[v] = true
		}
	}
}
