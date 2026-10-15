// Package cellmetrics defines the size of one terminal cell in CSS pixels.
// Layout, painting and ASCII sampling all share these values so that a
// viewport of N columns is N*WidthPx pixels wide and text lines are HeightPx
// pixels apart.
package cellmetrics

const (
	// WidthPx is the width of one terminal cell in CSS pixels.
	WidthPx = 10
	// HeightPx is the height of one terminal row in CSS pixels.
	HeightPx = 20
)
