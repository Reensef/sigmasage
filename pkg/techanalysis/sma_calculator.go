package techanalysis

import (
	"container/ring"
	"fmt"
)

// Rename to Averager???
type SMACalculator struct {
	windowSize int        // SMA window size
	windowSum  float64    // Current sum of values in the window
	r          *ring.Ring // Queue for storing values
}

func NewSMACalculator(windowSize int, initialData []float64) (
	*SMACalculator,
	error,
) {
	if len(initialData) != windowSize {
		return nil, fmt.Errorf("initial data length (%d) must match window size (%d)", len(initialData), windowSize)
	}

	sma := &SMACalculator{
		windowSize: windowSize,
		windowSum:  0.0,
		r:          ring.New(windowSize),
	}

	for _, value := range initialData {
		sma.r.Value = value
		sma.r = sma.r.Next()
		sma.windowSum += value
	}

	return sma, nil
}

func (sma *SMACalculator) Update(value float64) float64 {
	oldestValue := sma.r.Value.(float64)
	sma.windowSum -= oldestValue

	sma.r.Value = value
	sma.r = sma.r.Next()
	sma.windowSum += value

	return sma.windowSum / float64(sma.windowSize)
}
