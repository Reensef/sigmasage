package indicators

import (
	"container/ring"
)

type SMA struct {
	windowSize int        // Размер окна SMA
	windowSum  float64    // Текущая сумма значений в окне
	count      int        // Количество элементов в окне
	r          *ring.Ring // Очередь для хранения значений
}

// Создание нового SMA
func NewSMA(windowSize int) *SMA {
	return &SMA{
		windowSize: windowSize,
		windowSum:  0.0,
		count:      0,
		r:          ring.New(windowSize),
	}
}

// Обновление SMA новым значением
func (sma *SMA) Update(newValue float64) float64 {
	// Если окно заполнено, удаляем самое старое значение
	if sma.count >= sma.windowSize {
		oldestValue := sma.r.Value.(float64)
		sma.windowSum -= oldestValue
	} else {
		sma.count++
	}

	sma.r.Value = newValue
	sma.r = sma.r.Next()
	sma.windowSum += newValue

	return sma.windowSum / float64(sma.count)
}
