package ocr

import (
	"fmt"
	"image/png"
	"os"
	"testing"

	"github.com/disintegration/imaging"

	"sudokuSolver/internal/sudoku"
	"sudokuSolver/internal/vision"
)

func TestRecognizeGeneratedBoard(t *testing.T) {
	img, err := imaging.Open("../../sudoku_test.png")
	if err != nil {
		t.Skipf("test image not found: %v", err)
	}

	grid, err := vision.DetectGrid(img)
	if err != nil {
		t.Fatalf("detect grid: %v", err)
	}

	rec, err := NewRecognizer()
	if err != nil {
		t.Fatalf("ocr: %v", err)
	}

	expected := sudoku.Board{
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

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			cell := grid.Cells[r][c]
			empty := isEmpty(cell, 0.03)
			prepped := prepare(cell)
			_ = prepped

			digit, _, err := rec.RecognizeCell(cell)
			if err != nil {
				t.Errorf("cell (%d,%d): %v", r, c, err)
				continue
			}

			if r == 0 && c < 3 {
				name := fmt.Sprintf("../../debug_cell_%d_%d.png", r, c)
				f, _ := os.Create(name)
				if f != nil {
					png.Encode(f, prepare(cell))
					f.Close()
				}
			}

			want := expected.Get(r, c)
			if want == 0 {
				if !empty && digit != 0 {
					t.Errorf("cell (%d,%d): expected empty, got %d", r, c, digit)
				}
			} else if digit != want {
				t.Errorf("cell (%d,%d): expected %d, got %d (empty=%v)", r, c, want, digit, empty)
			}
		}
	}
}
