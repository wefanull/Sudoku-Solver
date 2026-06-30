package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"sudokuSolver/internal/sudoku"
)

func main() {
	const size = 900
	const cell = size / 9

	// A known Sudoku puzzle (solution is deterministic).
	board := sudoku.Board{
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

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	// Draw grid lines.
	for i := 0; i <= 9; i++ {
		thickness := size / 450
		if i%3 == 0 {
			thickness = size * 3 / 450
		}
		coord := i * cell
		for t := 0; t < thickness; t++ {
			drawLine(img, coord+t, 0, coord+t, size, color.Black)
			drawLine(img, 0, coord+t, size, coord+t, color.Black)
		}
	}

	// Draw digits using a system TrueType font.
	fontBytes, err := os.ReadFile(`C:\Windows\Fonts\calibri.ttf`)
	if err != nil {
		panic(err)
	}
	parsed, err := opentype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    64,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			v := board.Get(r, c)
			if v == 0 {
				continue
			}
			s := fmt.Sprintf("%d", v)
			bounds, _ := drawer.BoundString(s)
			w := (bounds.Max.X - bounds.Min.X).Ceil()
			h := (bounds.Max.Y - bounds.Min.Y).Ceil()
			x := c*cell + cell/2 - w/2
			y := r*cell + cell/2 + h/2
			drawer.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
			drawer.DrawString(s)
		}
	}

	out, err := os.Create("sudoku_test.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		panic(err)
	}
	fmt.Println("Generated sudoku_test.png")
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	if x1 == x2 {
		for y := y1; y < y2; y++ {
			img.Set(x1, y, col)
		}
	} else if y1 == y2 {
		for x := x1; x < x2; x++ {
			img.Set(x, y1, col)
		}
	}
}
