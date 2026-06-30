package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/disintegration/imaging"

	"sudokuSolver/internal/ocr"
	"sudokuSolver/internal/sudoku"
	"sudokuSolver/internal/vision"
)

func main() {
	start := time.Now()

	var (
		input   = flag.String("input", "", "Path to the Sudoku screenshot (png, jpg, etc.)")
		debug   = flag.String("debug", "", "Optional path to save the image with detected grid lines")
		help    = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help || *input == "" {
		fmt.Println("Usage: sudokusolver -input screenshot.png [-debug debug.png]")
		flag.PrintDefaults()
		os.Exit(0)
	}

	img, err := imaging.Open(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening image: %v\n", err)
		os.Exit(1)
	}

	grid, err := vision.DetectGrid(img)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting grid: %v\n", err)
		os.Exit(1)
	}

	if *debug != "" {
		if err := saveDebug(grid, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving debug image: %v\n", err)
		}
	}

	recognizer, err := ocr.NewRecognizer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing OCR: %v\n", err)
		os.Exit(1)
	}

	board, err := readBoard(grid.Cells, recognizer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading board: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Detected board:")
	fmt.Println(board.String())

	solved, ok := board.Solved()
	if !ok {
		fmt.Fprintln(os.Stderr, "Could not solve the detected board. Please check the OCR output above.")
		os.Exit(1)
	}

	fmt.Println("Solved board:")
	fmt.Println(solved.String())
	
	fmt.Printf("\nTiempo total de ejecución: %v\n", time.Since(start))
}

func readBoard(cells [][]image.Image, recognizer *ocr.Recognizer) (sudoku.Board, error) {
	var board sudoku.Board
	var wg sync.WaitGroup
	
	// Usamos un canal para capturar el primer error que ocurra en cualquier goroutine
	errChan := make(chan error, 81)

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			wg.Add(1)
			
			// Lanzamos una goroutine por cada celda
			go func(row, col int) {
				defer wg.Done()
				
				digit, empty, err := recognizer.RecognizeCell(cells[row][col])
				if err != nil {
					errChan <- fmt.Errorf("cell (%d,%d): %w", row, col, err)
					return
				}
				if !empty && digit >= 1 && digit <= 9 {
					// Es seguro escribir concurrentemente porque cada goroutine
					// escribe en una posición [row][col] única.
					board.Set(row, col, digit)
				}
			}(r, c)
		}
	}

	// Esperamos a que las 81 goroutines terminen
	wg.Wait()
	close(errChan)

	// Si alguna goroutine falló, retornamos el primer error
	if err, ok := <-errChan; ok {
		return board, err
	}

	return board, nil
}

func saveDebug(grid *vision.Grid, path string) error {
	out := vision.DebugLines(grid.Board, grid.HLines, grid.VLines)
	ext := filepath.Ext(path)
	if ext == "" {
		path += ".png"
	}
	return imaging.Save(out, path)
}
