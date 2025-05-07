package marketdata

import (
	"context"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"
)

// TinkoffMarketdata implements an observer that distributes candle data to subscribers
type MockMarketdata struct {
	mu                            sync.RWMutex
	candleSubscribers             map[MarketData][]chan Candle
	candlesNotifyCancel           context.CancelFunc
	isNotifyingCandlesSubscribers bool
}

func NewMockMarketdata() *MockMarketdata {
	return &MockMarketdata{
		candleSubscribers: make(map[MarketData][]chan Candle),
	}
}

// Subscribe returns a channel that will receive candle data for the specified instrument
func (t *MockMarketdata) SubscribeCandles(marketDataInfo MarketData) (<-chan Candle, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	ch := make(chan Candle, 100)

	if _, exists := t.candleSubscribers[marketDataInfo]; !exists {
		t.candleSubscribers[marketDataInfo] = make([]chan Candle, 0)
	}
	t.candleSubscribers[marketDataInfo] = append(t.candleSubscribers[marketDataInfo], ch)

	if !t.isNotifyingCandlesSubscribers {
		t.startNotifyingCandlesSubscribers(marketDataInfo)
		t.isNotifyingCandlesSubscribers = true
	}

	return ch, nil
}

// Unsubscribe removes a channel from the subscribers list
func (t *MockMarketdata) UnsubscribeCandles(marketDataInfo MarketData, ch <-chan Candle) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if subscribers, exists := t.candleSubscribers[marketDataInfo]; exists {
		for i, subscriber := range subscribers {
			if subscriber == ch {
				t.candleSubscribers[marketDataInfo] = slices.Delete(subscribers, i, i+1)
				close(subscriber)
				break
			}
		}

		if len(t.candleSubscribers[marketDataInfo]) == 0 {
			delete(t.candleSubscribers, marketDataInfo)
		}
	}
}

func (t *MockMarketdata) startNotifyingCandlesSubscribers(info MarketData) {
	var ctx context.Context
	ctx, t.candlesNotifyCancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func(ctx context.Context) {
		mockCandles := []Candle{
			{info, 100.0, 99.0, 100.0, 100.0, 100.0, time.Now(), time.Now()},
			{info, 102.0, 101.0, 102.0, 102.0, 102.0, time.Now(), time.Now()},
			{info, 101.0, 100.0, 101.0, 101.0, 101.0, time.Now(), time.Now()},
			{info, 103.0, 102.0, 103.0, 103.0, 103.0, time.Now(), time.Now()},
			{info, 98.0, 97.0, 98.0, 98.0, 98.0, time.Now(), time.Now()},
			{info, 99.0, 98.0, 99.0, 99.0, 99.0, time.Now(), time.Now()},
			{info, 101.5, 100.5, 101.5, 101.5, 101.5, time.Now(), time.Now()},
			{info, 102.5, 101.5, 102.5, 102.5, 102.5, time.Now(), time.Now()},
			{info, 100.5, 99.5, 100.5, 100.5, 100.5, time.Now(), time.Now()},
			{info, 101.8, 100.8, 101.8, 101.8, 101.8, time.Now(), time.Now()},
		}
		currentCandleIndex := 0
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				t.mu.RLock()
				/// ???
				for _, subscribers := range t.candleSubscribers {
					for _, subscriber := range subscribers {
						subscriber <- mockCandles[currentCandleIndex]
					}
				}
				t.mu.RUnlock()

				currentCandleIndex = (currentCandleIndex + 1) % len(mockCandles)
			}
		}
	}(ctx)
}
