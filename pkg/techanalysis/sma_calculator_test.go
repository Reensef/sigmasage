package techanalysis

import (
	"testing"
)

func TestNewSMACalculator(t *testing.T) {
	windowSize := 3
	initialData := []float64{1.0, 2.0, 3.0}
	sma, err := NewSMACalculator(windowSize, initialData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sma.windowSize != windowSize {
		t.Errorf("expected windowSize %d, got %d", windowSize, sma.windowSize)
	}
	expectedSum := 6.0
	if sma.windowSum != expectedSum {
		t.Errorf("expected windowSum %.2f, got %.2f", expectedSum, sma.windowSum)
	}
}

func TestNewSMACalculator_InvalidInitialData(t *testing.T) {
	windowSize := 3
	initialData := []float64{1.0, 2.0}
	_, err := NewSMACalculator(windowSize, initialData)
	if err == nil {
		t.Fatal("expected error for mismatched initial data length, got nil")
	}
}

func TestSMACalculator_Update(t *testing.T) {
	windowSize := 10
	initialData := []float64{10.11, 10.12, 10.13, 10.14, 10.15, 10.16, 10.17, 10.18, 10.19, 10.20}
	sma, _ := NewSMACalculator(windowSize, initialData)

	result := sma.Update(4.0)
	expected := (10.12 + 10.13 + 10.14 + 10.15 + 10.16 + 10.17 + 10.18 + 10.19 + 10.20 + 4.0) / 10.0
	if result != expected {
		t.Errorf("expected SMA %.2f, got %.2f", expected, result)
	}

	result = sma.Update(5.0)
	expected = (10.13 + 10.14 + 10.15 + 10.16 + 10.17 + 10.18 + 10.19 + 10.20 + 4.0 + 5.0) / 10.0
	if result != expected {
		t.Errorf("expected SMA %.2f, got %.2f", expected, result)
	}
}
