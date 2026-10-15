package cellmetrics

import "testing"

func TestCellMetrics(t *testing.T) {
	if WidthPx != 10 {
		t.Errorf("WidthPx = %d, want 10", WidthPx)
	}
	if HeightPx != 20 {
		t.Errorf("HeightPx = %d, want 20", HeightPx)
	}
}
