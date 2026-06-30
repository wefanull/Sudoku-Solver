package ocr

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/disintegration/imaging"
)

// Recognizer wraps the Tesseract OCR command.
type Recognizer struct {
	tesseract string
}

// NewRecognizer creates a Recognizer, locating the tesseract executable.
func NewRecognizer() (*Recognizer, error) {
	cmd := os.Getenv("TESSERACT_CMD")
	if cmd == "" {
		if runtime.GOOS == "windows" {
			cmd = `C:\Program Files\Tesseract-OCR\tesseract.exe`
		} else {
			cmd = "tesseract"
		}
	}

	if _, err := os.Stat(cmd); err != nil {
		path, lookErr := exec.LookPath("tesseract")
		if lookErr != nil {
			return nil, fmt.Errorf("tesseract not found: set TESSERACT_CMD or add tesseract to PATH: %w", lookErr)
		}
		cmd = path
	}
	return &Recognizer{tesseract: cmd}, nil
}

// RecognizeCell reads a single digit from a cell image.
// The second return value is true when the cell looks empty.
func (r *Recognizer) RecognizeCell(img image.Image) (int, bool, error) {
	if isEmpty(img, 0.03) {
		return 0, true, nil
	}

	prepped := prepare(img)

	tmpFile, err := os.CreateTemp("", "sudoku-cell-*.png")
	if err != nil {
		return 0, false, err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if err := png.Encode(tmpFile, prepped); err != nil {
		tmpFile.Close()
		return 0, false, err
	}
	tmpFile.Close()

	outBase := tmpName + ".out"
	defer os.Remove(outBase + ".txt")

	args := []string{
		tmpName,
		outBase,
		"--psm", "10",
		"-c", "tessedit_char_whitelist=0123456789",
		"quiet",
	}
	output, err := exec.Command(r.tesseract, args...).CombinedOutput()
	if err != nil {
		return 0, false, fmt.Errorf("tesseract failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	text, err := os.ReadFile(outBase + ".txt")
	if err != nil {
		return 0, false, err
	}

	digit := parseDigit(string(text))
	if digit == 0 {
		return 0, true, nil
	}
	return digit, false, nil
}

// prepare cleans up a cell image so Tesseract has the best chance of success.
func prepare(img image.Image) image.Image {
	gray := imaging.Grayscale(img)
	// Tesseract expects a light background and dark text.
	if estimateBackground(gray) < 128 {
		gray = imaging.Invert(gray)
	}
	// Resize to a larger, fixed size so the digit is not too small.
	resized := imaging.Fit(gray, 200, 200, imaging.Lanczos)
	// Add a white border so the digit does not touch the image edges.
	canvas := imaging.New(240, 240, image.White)
	offset := image.Pt((240-resized.Bounds().Dx())/2, (240-resized.Bounds().Dy())/2)
	padded := imaging.Paste(canvas, resized, offset)
	// Slight contrast boost to harden edges while keeping anti-aliasing.
	return imaging.AdjustContrast(padded, 20)
}

// estimateBackground samples the four corners to guess the background brightness.
func estimateBackground(img image.Image) uint8 {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	points := []image.Point{
		{bounds.Min.X + w/8, bounds.Min.Y + h/8},
		{bounds.Max.X - w/8, bounds.Min.Y + h/8},
		{bounds.Min.X + w/8, bounds.Max.Y - h/8},
		{bounds.Max.X - w/8, bounds.Max.Y - h/8},
	}
	var sum uint64
	for _, p := range points {
		sum += uint64(meanGray(img, p.X, p.Y))
	}
	return uint8(sum / 4)
}

// isEmpty returns true if the cell contains very few dark pixels.
func isEmpty(img image.Image, thresholdRatio float64) bool {
	gray := imaging.Grayscale(img)
	total := gray.Bounds().Dx() * gray.Bounds().Dy()
	if total == 0 {
		return true
	}
	dark := 0
	for y := gray.Bounds().Min.Y; y < gray.Bounds().Max.Y; y++ {
		for x := gray.Bounds().Min.X; x < gray.Bounds().Max.X; x++ {
			if meanGray(gray, x, y) < 180 {
				dark++
			}
		}
	}
	return float64(dark)/float64(total) < thresholdRatio
}

// binarize converts an image to black & white using the given threshold.
func binarize(img image.Image, threshold uint8) *image.NRGBA {
	bounds := img.Bounds()
	out := imaging.New(bounds.Dx(), bounds.Dy(), image.White)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if meanGray(img, x, y) < threshold {
				out.Set(x, y, image.Black)
			}
		}
	}
	return out
}

func meanGray(img image.Image, x, y int) uint8 {
	r, g, b, _ := img.At(x, y).RGBA()
	return uint8((r*299 + g*587 + b*114) / 1000 >> 8)
}

func meanGrayValue(img image.Image) uint8 {
	bounds := img.Bounds()
	var sum uint64
	n := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			sum += uint64(meanGray(img, x, y))
			n++
		}
	}
	if n == 0 {
		return 255
	}
	return uint8(sum / uint64(n))
}

var digitRe = regexp.MustCompile(`\d`)

func parseDigit(s string) int {
	s = strings.TrimSpace(s)
	m := digitRe.FindString(s)
	if m == "" {
		return 0
	}
	return int(m[0] - '0')
}

