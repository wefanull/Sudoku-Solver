package vision

import (
	"image"
	"image/color"
	"sort"

	"github.com/disintegration/imaging"
)

// Cell represents a single Sudoku cell image together with its board coordinates.
type Cell struct {
	Row, Col int
	Img      image.Image
}

// Grid holds the detected board region and the cropped cell images.
type Grid struct {
	Board  image.Image
	HLines []int // y coordinates of the 10 horizontal grid lines
	VLines []int // x coordinates of the 10 vertical grid lines
	Cells  [][]image.Image
}

// DetectGrid loads an image, finds the Sudoku board, detects the grid lines and
// returns the 81 cropped cells.
func DetectGrid(img image.Image) (*Grid, error) {
	gray := imaging.Grayscale(img)
	gray = autoCrop(gray)

	hLines, vLines, err := findGridLines(gray)
	if err != nil {
		return nil, err
	}

	cells := cropCells(gray, hLines, vLines)
	return &Grid{
		Board:  gray,
		HLines: hLines,
		VLines: vLines,
		Cells:  cells,
	}, nil
}

// autoCrop trims white borders and returns the smallest rectangle that contains
// the dark board lines.
func autoCrop(img *image.NRGBA) *image.NRGBA {
	bounds := img.Bounds()
	threshold := uint8(240)

	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if isDark(img, x, y, threshold) {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if minX >= maxX || minY >= maxY {
		return img
	}

	// Add a small margin so outer grid lines stay inside.
	margin := 2
	rect := image.Rect(minX-margin, minY-margin, maxX+margin+1, maxY+margin+1)
	rect = rect.Intersect(bounds)
	return imaging.Crop(img, rect)
}

func isDark(img *image.NRGBA, x, y int, threshold uint8) bool {
	r, g, b, _ := img.At(x, y).RGBA()
	// RGBA values are 16-bit; compare against threshold shifted.
	gray := uint8((r*299 + g*587 + b*114) / 1000 >> 8)
	return gray < threshold
}

func meanGray(img *image.NRGBA, x, y int) uint8 {
	r, g, b, _ := img.At(x, y).RGBA()
	return uint8((r*299 + g*587 + b*114) / 1000 >> 8)
}

// findGridLines detects 10 horizontal and 10 vertical grid lines by looking at
// the inverted grayscale projection of the board.
func findGridLines(img *image.NRGBA) ([]int, []int, error) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	threshold := otsuThreshold(img)

	horiz := make([]float64, h)
	vert := make([]float64, w)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			g := meanGray(img, bounds.Min.X+x, bounds.Min.Y+y)
			var inv float64
			if g < threshold {
				inv = float64(255 - g)
			}
			horiz[y] += inv
			vert[x] += inv
		}
	}

	hLines := findPeaks(horiz, 10)
	vLines := findPeaks(vert, 10)

	if len(hLines) != 10 || len(vLines) != 10 {
		return fallbackLines(w, h)
	}
	return hLines, vLines, nil
}

// fallbackLines returns evenly spaced lines when automatic detection fails.
func fallbackLines(w, h int) ([]int, []int, error) {
	hLines := make([]int, 10)
	vLines := make([]int, 10)
	for i := 0; i < 10; i++ {
		hLines[i] = i * h / 9
		vLines[i] = i * w / 9
	}
	return hLines, vLines, nil
}

// findPeaks returns the positions of the n strongest lines from a 1-D projection.
func findPeaks(proj []float64, n int) []int {
	maxVal := 0.0
	for _, v := range proj {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		return nil
	}

	threshold := maxVal * 0.20
	type peak struct {
		center int
		height float64
	}
	var peaks []peak

	inPeak := false
	var start int
	for i, v := range proj {
		if v >= threshold && !inPeak {
			inPeak = true
			start = i
		}
		if v < threshold && inPeak {
			center, height := centroid(proj, start, i)
			peaks = append(peaks, peak{center: center, height: height})
			inPeak = false
		}
	}
	if inPeak {
		center, height := centroid(proj, start, len(proj))
		peaks = append(peaks, peak{center: center, height: height})
	}

	if len(peaks) < n {
		return nil
	}

	sort.Slice(peaks, func(i, j int) bool {
		return peaks[i].height > peaks[j].height
	})
	peaks = peaks[:n]

	sort.Slice(peaks, func(i, j int) bool {
		return peaks[i].center < peaks[j].center
	})

	result := make([]int, n)
	for i, p := range peaks {
		result[i] = p.center
	}
	return result
}

// centroid calculates the weighted center and max height of a peak region.
func centroid(proj []float64, start, end int) (int, float64) {
	var sum, weighted float64
	maxVal := 0.0
	for i := start; i < end && i < len(proj); i++ {
		v := proj[i]
		weighted += float64(i) * v
		sum += v
		if v > maxVal {
			maxVal = v
		}
	}
	if sum == 0 {
		return (start + end) / 2, 0
	}
	return int(weighted/sum + 0.5), maxVal
}

// cropCells extracts the 9x9 cells between the detected grid lines.
func cropCells(img *image.NRGBA, hLines, vLines []int) [][]image.Image {
	cells := make([][]image.Image, 9)
	for r := 0; r < 9; r++ {
		cells[r] = make([]image.Image, 9)
		y1 := hLines[r]
		y2 := hLines[r+1]
		for c := 0; c < 9; c++ {
			x1 := vLines[c]
			x2 := vLines[c+1]

			// Shrink the crop slightly towards the centre to avoid grid lines.
			insetX := (x2 - x1) / 10
			insetY := (y2 - y1) / 10
			rect := image.Rect(x1+insetX, y1+insetY, x2-insetX, y2-insetY)
			cells[r][c] = imaging.Crop(img, rect)
		}
	}
	return cells
}

// otsuThreshold computes a binary threshold using Otsu's method.
func otsuThreshold(img *image.NRGBA) uint8 {
	bounds := img.Bounds()
	var histogram [256]int
	count := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			g := meanGray(img, x, y)
			histogram[g]++
			count++
		}
	}
	if count == 0 {
		return 128
	}

	var sum float64
	for i, v := range histogram {
		sum += float64(i * v)
	}

	var sumB float64
	wB := 0
	maxVariance := 0.0
	threshold := uint8(0)

	for t := 0; t < 256; t++ {
		wB += histogram[t]
		if wB == 0 {
			continue
		}
		wF := count - wB
		if wF == 0 {
			break
		}
		sumB += float64(t * histogram[t])
		mB := sumB / float64(wB)
		mF := (sum - sumB) / float64(wF)
		variance := float64(wB) * float64(wF) * (mB - mF) * (mB - mF)
		if variance > maxVariance {
			maxVariance = variance
			threshold = uint8(t)
		}
	}
	return threshold
}

// DebugLines draws the detected grid lines on a copy of the board image.
func DebugLines(img image.Image, hLines, vLines []int) image.Image {
	out := imaging.Clone(img)
	bounds := out.Bounds()
	red := color.NRGBA{R: 255, A: 255}
	for _, y := range hLines {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			out.Set(x, bounds.Min.Y+y, red)
		}
	}
	for _, x := range vLines {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			out.Set(bounds.Min.X+x, y, red)
		}
	}
	return out
}


