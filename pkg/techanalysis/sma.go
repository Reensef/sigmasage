package techanalysis

import (
	"container/ring"
	"fmt"
)

type SMACalculator struct {
	windowSize int        // Размер окна SMA
	windowSum  float64    // Текущая сумма значений в окне
	r          *ring.Ring // Очередь для хранения значений
}

// Создание нового SMA
func NewSMACalculator(windowSize int, initialData []float64) (*SMACalculator, error) {
	if len(initialData) != windowSize {
		return nil, fmt.Errorf("initial data length (%d) must match window size (%d)", len(initialData), windowSize)
	}

	sma := &SMACalculator{
		windowSize: windowSize,
		windowSum:  0.0,
		r:          ring.New(windowSize),
	}

	// Initialize the ring with initial data
	for _, value := range initialData {
		sma.r.Value = value
		sma.r = sma.r.Next()
		sma.windowSum += value
	}

	return sma, nil
}

// Обновление SMA новым значением
func (sma *SMACalculator) Update(value float64) float64 {
	// Удаляем старое значение
	oldestValue := sma.r.Value.(float64)
	sma.windowSum -= oldestValue

	// Добавляем новое значение
	sma.r.Value = value
	sma.r = sma.r.Next()
	sma.windowSum += value

	return sma.windowSum / float64(sma.windowSize)
}
