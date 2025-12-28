package recorder

import (
	"fmt"

	"github.com/kbinani/screenshot"
)

// GetDisplayBounds returns the width and height of the primary display
// reducePercent is the percentage to reduce the size (e.g., 30 means 30% of original size)
func GetDisplayBounds(reducePercent int) (int, int, error) {
	// Get the number of active displays
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return 0, 0, fmt.Errorf("no active displays found")
	}

	// Get the bounds of the primary display (index 0)
	bounds := screenshot.GetDisplayBounds(0)

	// Calculate reduced dimensions
	width := bounds.Dx() * reducePercent / 100
	height := bounds.Dy() * reducePercent / 100

	// If reduced dimensions are too small, return original resolution
	if width < 640 || height < 480 {
		return bounds.Dx(), bounds.Dy(), nil
	}

	// Return reduced width and height
	return width, height, nil
}
