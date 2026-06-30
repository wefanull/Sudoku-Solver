package sudoku

import (
	"fmt"
	"strings"
)

// Board represents a 9x9 Sudoku grid. 0 means empty.
type Board [9][9]int

// NewBoard creates an empty board.
func NewBoard() Board {
	return Board{}
}

// Set assigns a value to a cell.
func (b *Board) Set(row, col, value int) {
	b[row][col] = value
}

// Get returns the value of a cell.
func (b Board) Get(row, col int) int {
	return b[row][col]
}

// Valid checks if placing num at (row, col) is legal.
func (b Board) Valid(row, col, num int) bool {
	for i := 0; i < 9; i++ {
		if b[row][i] == num {
			return false
		}
		if b[i][col] == num {
			return false
		}
	}
	startRow := (row / 3) * 3
	startCol := (col / 3) * 3
	for r := startRow; r < startRow+3; r++ {
		for c := startCol; c < startCol+3; c++ {
			if b[r][c] == num {
				return false
			}
		}
	}
	return true
}

// Solve solves the board in-place using backtracking.
// Returns true if a solution was found.
func (b *Board) Solve() bool {
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			if b[row][col] != 0 {
				continue
			}
			for num := 1; num <= 9; num++ {
				if !b.Valid(row, col, num) {
					continue
				}
				b[row][col] = num
				if b.Solve() {
					return true
				}
				b[row][col] = 0
			}
			return false
		}
	}
	return true
}

// Solved returns a copy of the board solved, leaving the original untouched.
func (b Board) Solved() (Board, bool) {
	copy := b
	if !copy.Solve() {
		return b, false
	}
	return copy, true
}

// String returns a human-readable representation of the board.
func (b Board) String() string {
	var sb strings.Builder
	for row := 0; row < 9; row++ {
		if row%3 == 0 && row != 0 {
			sb.WriteString("------+-------+------\n")
		}
		for col := 0; col < 9; col++ {
			if col%3 == 0 && col != 0 {
				sb.WriteString("| ")
			}
			v := b[row][col]
			if v == 0 {
				sb.WriteString(". ")
			} else {
				sb.WriteString(fmt.Sprintf("%d ", v))
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}
